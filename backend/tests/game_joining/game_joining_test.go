package game_joining

import (
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
	r.POST("/join-game/:id", game.JoinGameHandler(db))
	return r
}

func TestJoinGame(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := tests.SetupTestDB(t)
		r := setupRouter(db)

		// Creating initial game for joining
		newGame := game.Game{Name: "Game to Join - " + uuid.New().String()}
		require.NoError(t, db.Create(&newGame).Error)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/join-game/"+newGame.ID.String(), nil)
		req.Header.Set("User-ID", uuid.New().String()) // Use a new UUID for the user
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Joined game")
	})

	t.Run("game not found", func(t *testing.T) {
		db := tests.SetupTestDB(t)
		r := setupRouter(db)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/join-game/"+uuid.New().String(), nil)
		req.Header.Set("User-ID", uuid.New().String())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "game not found")
	})

	t.Run("already in game", func(t *testing.T) {
		db := tests.SetupTestDB(t)
		r := setupRouter(db)

		// Creating initial game and player
			user := game.User{Email: "test@example.com"}
	require.NoError(t, db.Create(&user).Error)
	newGame := game.Game{Name: "Game to Join - " + uuid.New().String()}
	require.NoError(t, db.Create(&newGame).Error)
	require.NoError(t, db.Create(&game.Player{UserID: user.ID, GameID: newGame.ID}).Error)


		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/join-game/"+newGame.ID.String(), nil)
		req.Header.Set("User-ID", user.ID.String())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "already a participant")
	})

	t.Run("missing authentication", func(t *testing.T) {
		db := tests.SetupTestDB(t)
		r := setupRouter(db)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/join-game/"+uuid.New().String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "authentication required")
	})

	t.Run("turn order assignment", func(t *testing.T) {
		db := tests.SetupTestDB(t)
		r := setupRouter(db)

		user1 := game.User{Email: "player1@example.com"}
		require.NoError(t, db.Create(&user1).Error)
		user2 := game.User{Email: "player2@example.com"}
		require.NoError(t, db.Create(&user2).Error)
		user3 := game.User{Email: "player3@example.com"}
		require.NoError(t, db.Create(&user3).Error)

		newGame := game.Game{Name: "Turn Order Game - " + uuid.New().String()}
		require.NoError(t, db.Create(&newGame).Error)

		// Player 1 joins
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("POST", "/join-game/"+newGame.ID.String(), nil)
		req1.Header.Set("User-ID", user1.ID.String())
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Player 2 joins
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/join-game/"+newGame.ID.String(), nil)
		req2.Header.Set("User-ID", user2.ID.String())
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		// Player 3 joins
		w3 := httptest.NewRecorder()
		req3, _ := http.NewRequest("POST", "/join-game/"+newGame.ID.String(), nil)
		req3.Header.Set("User-ID", user3.ID.String())
		r.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusOK, w3.Code)

		// Verify turn orders
		var players []game.Player
		require.NoError(t, db.Where("game_id = ?", newGame.ID).Order("turn_order ASC").Find(&players).Error)
		assert.Len(t, players, 3)
		assert.Equal(t, user1.ID, players[0].UserID)
		assert.Equal(t, 0, players[0].TurnOrder)
		assert.Equal(t, user2.ID, players[1].UserID)
		assert.Equal(t, 1, players[1].TurnOrder)
		assert.Equal(t, user3.ID, players[2].UserID)
		assert.Equal(t, 2, players[2].TurnOrder)
	})
}
