package game_details

import (
	"encoding/json"
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
	r.GET("/games/:id", game.GetGameHandler(db))
	return r
}

func TestGetGame(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := tests.SetupTestDB(t)
		r := setupRouter(db)

		user := game.User{Email: "creator@example.com"}
		require.NoError(t, db.Create(&user).Error)
		newGame := game.Game{Name: "Test Game - " + uuid.New().String(), CreatorID: user.ID}
		require.NoError(t, db.Create(&newGame).Error)

		player1 := game.Player{UserID: user.ID, GameID: newGame.ID, TurnOrder: 0}
		require.NoError(t, db.Create(&player1).Error)

		// Re-fetch the game to ensure players are loaded
		var fetchedGame game.Game
		require.NoError(t, db.Preload("Players").First(&fetchedGame, "id = ?", newGame.ID).Error)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/games/"+fetchedGame.ID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		t.Logf("Response Body: %s", w.Body.String())

		var responseGame game.Game
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &responseGame))

		assert.Equal(t, fetchedGame.ID, responseGame.ID)
		assert.Equal(t, fetchedGame.Name, responseGame.Name)
		assert.Equal(t, fetchedGame.CreatorID, responseGame.CreatorID)
		assert.Len(t, responseGame.Players, 1)
		assert.Equal(t, player1.ID, responseGame.Players[0].ID)
	})

	t.Run("game not found", func(t *testing.T) {
		db := tests.SetupTestDB(t)
		r := setupRouter(db)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/games/"+uuid.New().String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "game not found")
	})
}

