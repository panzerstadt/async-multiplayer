package saves_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func setupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.POST("/games/:id/saves", game.UploadSaveHandler(db))
	r.GET("/games/:id/saves/latest", game.GetLatestSaveHandler(db))
	return r
}

func TestGetLatestSave(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	db.AutoMigrate(&game.User{}, &game.Game{}, &game.Player{}, &game.Save{})

	r := setupRouter(db)

	t.Run("successful case - 200 with correct Content-Disposition and binary body", func(t *testing.T) {
		user := &game.User{Email: "test@example.com"}
		db.Create(user)
		newGame := &game.Game{Name: "Test Game", CreatorID: user.ID}
		db.Create(newGame)
		db.Create(&game.Player{UserID: user.ID, GameID: newGame.ID})

		// Create a temporary file and save
		saveDir := fmt.Sprintf("saves/%s", newGame.ID)
		os.MkdirAll(saveDir, 0755)
		filePath := filepath.Join(saveDir, "latest_save.zip")
		zipContent, err := helpers.CreateDummyZip()
		require.NoError(t, err)
		err = os.WriteFile(filePath, zipContent.Bytes(), 0644)
		require.NoError(t, err)
		defer os.RemoveAll("saves")

		db.Create(&game.Save{GameID: newGame.ID, FilePath: filePath, UploadedBy: user.ID})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/games/"+newGame.ID.String()+"/saves/latest", nil)
		req.Header.Set("User-ID", user.ID.String())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=")
		assert.NotEmpty(t, w.Body.String())
	})

	t.Run("no saves yet - 404", func(t *testing.T) {
		user := &game.User{Email: "no-saves@example.com"}
		db.Create(user)
		newGame := &game.Game{Name: "No Saves Game", CreatorID: user.ID}
		db.Create(newGame)
		db.Create(&game.Player{UserID: user.ID, GameID: newGame.ID})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/games/"+newGame.ID.String()+"/saves/latest", nil)
		req.Header.Set("User-ID", user.ID.String())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("game not found - 404", func(t *testing.T) {
		user := &game.User{Email: "game-not-found@example.com"}
		db.Create(user)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/games/"+uuid.New().String()+"/saves/latest", nil)
		req.Header.Set("User-ID", user.ID.String())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("unauthorized - 403", func(t *testing.T) {
		user := &game.User{Email: "unauthorized@example.com"}
		db.Create(user)
		newGame := &game.Game{Name: "Unauthorized Game", CreatorID: user.ID}
		db.Create(newGame)

		// Note: user is not a player in this game

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/games/"+newGame.ID.String()+"/saves/latest", nil)
		req.Header.Set("User-ID", user.ID.String())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
