package saves_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/helpers"
	"panzerstadt/async-multiplayer/tests"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadSave(t *testing.T) {
	db, r, cfg, err := tests.SetupTestEnvironment()
	require.NoError(t, err)
	defer tests.TeardownTestEnvironment(db)

	t.Run("successful upload - 201", func(t *testing.T) {
		user, err := tests.CreateTestUser(db, "upload-success@example.com")
		require.NoError(t, err)
		newGame := &game.Game{Name: "Upload Success Game", CreatorID: user.ID}
		db.Create(newGame)
		db.Create(&game.Player{UserID: user.ID, GameID: newGame.ID})
		defer os.RemoveAll("saves")

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		zipContent, err := helpers.CreateDummyZip()
		require.NoError(t, err)
		part, _ := writer.CreateFormFile("file", "test.zip")
		part.Write(zipContent.Bytes())
		writer.Close()

		token, err := tests.GetTestUserToken(user.ID, user.Email, cfg)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/games/"+newGame.ID.String()+"/saves", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("game not found - 404", func(t *testing.T) {
		user, err := tests.CreateTestUser(db, "upload-game-not-found@example.com")
		require.NoError(t, err)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		zipContent, err := helpers.CreateDummyZip()
		require.NoError(t, err)
		part, _ := writer.CreateFormFile("file", "test.zip")
		part.Write(zipContent.Bytes())
		writer.Close()

		token, err := tests.GetTestUserToken(user.ID, user.Email, cfg)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/games/"+uuid.New().String()+"/saves", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("user not in game - 403", func(t *testing.T) {
		user, err := tests.CreateTestUser(db, "upload-not-in-game@example.com")
		require.NoError(t, err)
		newGame := &game.Game{Name: "Not In Game", CreatorID: user.ID}
		db.Create(newGame)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		zipContent, err := helpers.CreateDummyZip()
		require.NoError(t, err)
		part, _ := writer.CreateFormFile("file", "test.zip")
		part.Write(zipContent.Bytes())
		writer.Close()

		token, err := tests.GetTestUserToken(user.ID, user.Email, cfg)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/games/"+newGame.ID.String()+"/saves", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
