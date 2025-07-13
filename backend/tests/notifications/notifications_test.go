package notifications

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	socketio "github.com/googollee/go-socket.io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/helpers"
	"panzerstadt/async-multiplayer/tests"
)

func setupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	server := socketio.NewServer(nil)
	// Group save-related routes
	savesGroup := r.Group("/games/:id/saves")
	savesGroup.POST("", game.UploadSaveHandler(db, server))
	return r
}

func TestNotificationOnSaveUpload(t *testing.T) {
	db := tests.SetupTestDB(t)
	r := setupRouter(db)

	user1 := game.User{Email: "player1@example.com"}
	require.NoError(t, db.Create(&user1).Error)
	user2 := game.User{Email: "player2@example.com"}
	require.NoError(t, db.Create(&user2).Error)

	newGame := game.Game{Name: "Notification Game - " + uuid.New().String()}
	require.NoError(t, db.Create(&newGame).Error)

	require.NoError(t, db.Create(&game.Player{UserID: user1.ID, GameID: newGame.ID, TurnOrder: 0}).Error)
	require.NoError(t, db.Create(&game.Player{UserID: user2.ID, GameID: newGame.ID, TurnOrder: 1}).Error)

	// Simulate file upload by user1
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	zipContent, err := helpers.CreateDummyZip()
	require.NoError(t, err)
	part, _ := writer.CreateFormFile("file", "test.zip")
	part.Write(zipContent.Bytes())
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/games/"+newGame.ID.String()+"/saves", body)
	req.Header.Set("User-ID", user1.ID.String())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	// In a real test, you would assert that a notification was sent (e.g., by mocking the notifier)
}
