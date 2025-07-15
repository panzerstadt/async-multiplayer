package notifications

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/helpers"
	"panzerstadt/async-multiplayer/tests"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationOnSaveUpload(t *testing.T) {
	mockNotifier := tests.NewMockNotifier()
	db, r, cfg := tests.SetupTestEnvironmentWithNotifier(t, mockNotifier)
	defer tests.TeardownTestEnvironment(db)

	user1, err := tests.CreateTestUser(db, "player1@example.com")
	require.NoError(t, err)
	user2, err := tests.CreateTestUser(db, "player2@example.com")
	require.NoError(t, err)

	newGame := &game.Game{Name: "Notification Game - " + uuid.New().String(), CreatorID: user1.ID}
	require.NoError(t, db.Create(newGame).Error)

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

	token, err := tests.GetTestUserToken(user1.ID, user1.Email, cfg)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/games/"+newGame.ID.String()+"/saves", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "player2@example.com", mockNotifier.LastRecipientEmail)
	assert.Contains(t, mockNotifier.LastSubject, fmt.Sprintf("New save uploaded for game %s!", newGame.Name))
}
