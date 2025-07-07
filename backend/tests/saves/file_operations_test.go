package saves_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/helpers"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	db.AutoMigrate(&game.User{}, &game.Game{}, &game.Player{}, &game.Save{})
	return db
}

func setupFileOpRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.POST("/games/:id/saves", game.UploadSaveHandler(db))
	return r
}

func TestUploadSave(t *testing.T) {
	db := setupTestDB(t)
	r := setupFileOpRouter(db)

	t.Run("successful upload - 201", func(t *testing.T) {
		user := &game.User{Email: "upload-success@example.com"}
		db.Create(user)
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

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/games/"+newGame.ID.String()+"/saves", body)
		req.Header.Set("User-ID", user.ID.String())
		req.Header.Set("Content-Type", writer.FormDataContentType())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("game not found - 404", func(t *testing.T) {
		db := setupTestDB(t)
		r := setupFileOpRouter(db)

		user := &game.User{Email: "upload-game-not-found@example.com"}
		db.Create(user)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		zipContent, err := helpers.CreateDummyZip()
		require.NoError(t, err)
		part, _ := writer.CreateFormFile("file", "test.zip")
		part.Write(zipContent.Bytes())
		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/games/"+uuid.New().String()+"/saves", body)
		req.Header.Set("User-ID", user.ID.String())
		req.Header.Set("Content-Type", writer.FormDataContentType())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("user not in game - 403", func(t *testing.T) {
		db := setupTestDB(t)
		r := setupFileOpRouter(db)

		user := &game.User{Email: "upload-not-in-game@example.com"}
		db.Create(user)
		newGame := &game.Game{Name: "Not In Game", CreatorID: user.ID}
		db.Create(newGame)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		zipContent, err := helpers.CreateDummyZip()
		require.NoError(t, err)
		part, _ := writer.CreateFormFile("file", "test.zip")
		part.Write(zipContent.Bytes())
		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/games/"+newGame.ID.String()+"/saves", body)
		req.Header.Set("User-ID", user.ID.String())
		req.Header.Set("Content-Type", writer.FormDataContentType())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

