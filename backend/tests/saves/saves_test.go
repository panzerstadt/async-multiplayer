package saves_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/helpers"
	"panzerstadt/async-multiplayer/tests"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLatestSave(t *testing.T) {
	db, r, cfg, err := tests.SetupTestEnvironment()
	require.NoError(t, err)
	defer tests.TeardownTestEnvironment(db)

	t.Run("successful case - 200 with correct Content-Disposition and binary body", func(t *testing.T) {
		user, err := tests.CreateTestUser(db, "test@example.com")
		require.NoError(t, err)
		newGame := &game.Game{Name: "Test Game", CreatorID: user.ID}
		db.Create(newGame)
		db.Create(&game.Player{UserID: user.ID, GameID: newGame.ID})

		// Create a temporary file and save
		saveDir := fmt.Sprintf("../../saves/%s", newGame.ID)
		os.MkdirAll(saveDir, 0755)
		filePath := filepath.Join(saveDir, "latest_save.zip")
		zipContent, err := helpers.CreateDummyZip()
		require.NoError(t, err)
		err = os.WriteFile(filePath, zipContent.Bytes(), 0644)
		require.NoError(t, err)
		defer os.RemoveAll("saves")

		db.Create(&game.Save{GameID: newGame.ID, FilePath: filePath, UploadedBy: user.ID})

		token, err := tests.GetTestUserToken(user.ID, user.Email, cfg)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/games/"+newGame.ID.String()+"/saves/latest", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=")
		assert.NotEmpty(t, w.Body.String())
	})

	t.Run("no saves yet - 404", func(t *testing.T) {
		user, err := tests.CreateTestUser(db, "no-saves@example.com")
		require.NoError(t, err)
		newGame := &game.Game{Name: "No Saves Game", CreatorID: user.ID}
		db.Create(newGame)
		db.Create(&game.Player{UserID: user.ID, GameID: newGame.ID})

		token, err := tests.GetTestUserToken(user.ID, user.Email, cfg)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/games/"+newGame.ID.String()+"/saves/latest", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("game not found - 404", func(t *testing.T) {
		user, err := tests.CreateTestUser(db, "game-not-found@example.com")
		require.NoError(t, err)

		token, err := tests.GetTestUserToken(user.ID, user.Email, cfg)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/games/"+uuid.New().String()+"/saves/latest", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("unauthorized - 403", func(t *testing.T) {
		user, err := tests.CreateTestUser(db, "unauthorized@example.com")
		require.NoError(t, err)
		newGame := &game.Game{Name: "Unauthorized Game", CreatorID: user.ID}
		db.Create(newGame)

		// Note: user is not a player in this game
		token, err := tests.GetTestUserToken(user.ID, user.Email, cfg)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/games/"+newGame.ID.String()+"/saves/latest", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
