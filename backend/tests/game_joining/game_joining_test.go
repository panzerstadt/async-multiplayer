package game_joining

import (
	"net/http"
	"net/http/httptest"
	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/tests"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinGame(t *testing.T) {
	db, r, err := tests.SetupTestEnvironment()
	require.NoError(t, err)
	defer tests.TeardownTestEnvironment(db)

	t.Run("success", func(t *testing.T) {
		user, err := tests.CreateTestUser(db, "join-success@example.com")
		require.NoError(t, err)
		token, err := tests.GetTestUserToken(user.ID, user.Email)
		require.NoError(t, err)

		// Creating initial game for joining
		newGame := game.Game{Name: "Game to Join - " + uuid.New().String()}
		require.NoError(t, db.Create(&newGame).Error)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/join-game/"+newGame.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Joined game")
	})

	t.Run("game not found", func(t *testing.T) {
		user, err := tests.CreateTestUser(db, "join-notfound@example.com")
		require.NoError(t, err)
		token, err := tests.GetTestUserToken(user.ID, user.Email)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/join-game/"+uuid.New().String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "game not found")
	})

	t.Run("already in game", func(t *testing.T) {
		user, err := tests.CreateTestUser(db, "join-already@example.com")
		require.NoError(t, err)
		token, err := tests.GetTestUserToken(user.ID, user.Email)
		require.NoError(t, err)

		// Creating initial game and player
		newGame := game.Game{Name: "Game to Join - " + uuid.New().String()}
		require.NoError(t, db.Create(&newGame).Error)
		require.NoError(t, db.Create(&game.Player{UserID: user.ID, GameID: newGame.ID}).Error)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/join-game/"+newGame.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "already a participant")
	})

	t.Run("missing authentication", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/join-game/"+uuid.New().String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header is missing")
	})

	t.Run("turn order assignment", func(t *testing.T) {
		user1, err := tests.CreateTestUser(db, "player1@example.com")
		require.NoError(t, err)
		token1, err := tests.GetTestUserToken(user1.ID, user1.Email)
		require.NoError(t, err)

		user2, err := tests.CreateTestUser(db, "player2@example.com")
		require.NoError(t, err)
		token2, err := tests.GetTestUserToken(user2.ID, user2.Email)
		require.NoError(t, err)

		user3, err := tests.CreateTestUser(db, "player3@example.com")
		require.NoError(t, err)
		token3, err := tests.GetTestUserToken(user3.ID, user3.Email)
		require.NoError(t, err)

		newGame := game.Game{Name: "Turn Order Game - " + uuid.New().String()}
		require.NoError(t, db.Create(&newGame).Error)

		// Player 1 joins
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("POST", "/join-game/"+newGame.ID.String(), nil)
		req1.Header.Set("Authorization", "Bearer "+token1)
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Player 2 joins
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/join-game/"+newGame.ID.String(), nil)
		req2.Header.Set("Authorization", "Bearer "+token2)
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		// Player 3 joins
		w3 := httptest.NewRecorder()
		req3, _ := http.NewRequest("POST", "/join-game/"+newGame.ID.String(), nil)
		req3.Header.Set("Authorization", "Bearer "+token3)
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
