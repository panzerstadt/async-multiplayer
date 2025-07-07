package game_creation

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/tests"
)

func setupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.POST("/create-game", game.CreateGameHandler(db))
	return r
}

func TestCreateGame(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := tests.SetupTestDB(t)
		r := setupRouter(db)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/create-game", bytes.NewBufferString(`{"name":"Game 1"}`))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Game created")
	})

	t.Run("name already exists", func(t *testing.T) {
		db := tests.SetupTestDB(t)
		r := setupRouter(db)

		// Creating initial game with name
		require.NoError(t, db.Create(&game.Game{Name: "Game 1 - " + uuid.New().String()}).Error)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/create-game", bytes.NewBufferString(`{"name":"Game 1"}`))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "name already exists")
	})

	t.Run("missing name", func(t *testing.T) {
		db := tests.SetupTestDB(t)
		r := setupRouter(db)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/create-game", bytes.NewBufferString(`{}`))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "name is required")
	})
}

