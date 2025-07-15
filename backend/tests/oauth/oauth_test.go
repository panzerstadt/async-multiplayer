package oauth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/tests"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleOAuth(t *testing.T) {
	// Set up environment variables for OAuth configuration
	os.Setenv("GOOGLE_OAUTH_CLIENT_ID", "test-client-id")
	os.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "test-client-secret")
	os.Setenv("GOOGLE_OAUTH_REDIRECT_URL", "http://localhost:8080/auth/google/callback")

	db, r, cfg, err := tests.SetupTestEnvironment()
	require.NoError(t, err)
	defer tests.TeardownTestEnvironment(db)
	game.InitOAuth(cfg)

	t.Run("Google Login Redirect", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/auth/google/login", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "accounts.google.com/o/oauth2/auth")
		assert.Contains(t, w.Header().Get("Set-Cookie"), "oauthstate")
	})

	t.Run("Google Callback Handler", func(t *testing.T) {
		// Mock Google API server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id": "12345", "email": "test@example.com"}`))
		}))
		defer server.Close()

		// Override the Google API URL to point to our mock server
		originalURL := game.GetOAuthGoogleUrlAPI()
		game.SetOAuthGoogleUrlAPI(server.URL)
		defer func() { game.SetOAuthGoogleUrlAPI(originalURL) }()

		// Set up a mock OAuth2 server
		oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token": "test-token", "token_type": "Bearer"}`))
		}))
		defer oauthServer.Close()

		// Override the OAuth2 token URL
		originalEndpoint := game.GetOAuthConf().Endpoint
		game.GetOAuthConf().Endpoint.TokenURL = oauthServer.URL
		defer func() { game.GetOAuthConf().Endpoint = originalEndpoint }()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/auth/google/callback?state=test-state&code=test-code", nil)
		req.AddCookie(&http.Cookie{Name: "oauthstate", Value: "test-state"})
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		redirectURL, err := w.Result().Location()
		require.NoError(t, err)
		assert.Contains(t, redirectURL.String(), "token=")

		// Verify that a user was created in the database
		var user game.User
		result := db.Where("email = ?", "test@example.com").First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, "test@example.com", user.Email)
	})
}

