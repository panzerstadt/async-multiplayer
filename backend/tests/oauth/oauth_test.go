package oauth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/tests"
)

func TestGoogleOAuth(t *testing.T) {
	// Set up environment variables for OAuth configuration
	os.Setenv("GOOGLE_OAUTH_CLIENT_ID", "test-client-id")
	os.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "test-client-secret")
	os.Setenv("GOOGLE_OAUTH_REDIRECT_URL", "http://localhost:8080/auth/google/callback")

	game.InitOAuth()

	db := tests.SetupTestDB(t)
	r := tests.SetupRouter(db)

	t.Run("Google Login Redirect", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/auth/google/login", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "accounts.google.com/o/oauth2/auth")
		assert.Contains(t, w.Header().Get("Set-Cookie"), "oauthstate")
	})

	// TODO: Add test for Google Callback Handler (requires mocking external HTTP calls)
}

