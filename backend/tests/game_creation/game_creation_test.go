package game_creation

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/tests"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGame(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, r, cfg, err := tests.SetupTestEnvironment()
		require.NoError(t, err)
		defer tests.TeardownTestEnvironment(db)
		user, err := tests.CreateTestUser(db, "creategame-success@example.com")
		require.NoError(t, err)
		token, err := tests.GetTestUserToken(user.ID, user.Email, cfg)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/create-game", bytes.NewBufferString(`{"name":"Game 1"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Game created")
	})

	t.Run("name already exists", func(t *testing.T) {
		db, r, cfg, err := tests.SetupTestEnvironment()
		require.NoError(t, err)
		defer tests.TeardownTestEnvironment(db)

		user, err := tests.CreateTestUser(db, "creategame-exists@example.com")
		require.NoError(t, err)
		token, err := tests.GetTestUserToken(user.ID, user.Email, cfg)
		require.NoError(t, err)

		// Creating initial game with name
		require.NoError(t, db.Create(&game.Game{Name: "Game 1", CreatorID: user.ID}).Error)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/create-game", bytes.NewBufferString(`{"name":"Game 1"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "A game with this name already exists.")
	})

	t.Run("missing name", func(t *testing.T) {
		db, r, cfg, err := tests.SetupTestEnvironment()
		require.NoError(t, err)
		defer tests.TeardownTestEnvironment(db)

		user, err := tests.CreateTestUser(db, "creategame-missing@example.com")
		require.NoError(t, err)
		token, err := tests.GetTestUserToken(user.ID, user.Email, cfg)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/create-game", bytes.NewBufferString(`{}`))
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "name is required")
	})
}
