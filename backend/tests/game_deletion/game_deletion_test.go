package game_deletion_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/tests"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteGame(t *testing.T) {
	db, router, err := tests.SetupTestEnvironment()
	assert.NoError(t, err)
	defer tests.TeardownTestEnvironment(db)

	// 1. Create a user and a game
	creator, err := tests.CreateTestUser(db, "creator@test.com")
	assert.NoError(t, err)
	gameToCreate := game.Game{Name: "GameToDelete", CreatorID: creator.ID}
	db.Create(&gameToCreate)

	// 2. Create a save file for the game
	saveDir := fmt.Sprintf("../../saves/%s", gameToCreate.ID)
	os.MkdirAll(saveDir, 0755)
	saveFilePath := fmt.Sprintf("%s/test_save.zip", saveDir)
	os.Create(saveFilePath)
	saveRecord := game.Save{GameID: gameToCreate.ID, FilePath: saveFilePath, UploadedBy: creator.ID}
	db.Create(&saveRecord)

	// 3. Create a player for the game
	player := game.Player{UserID: creator.ID, GameID: gameToCreate.ID}
	db.Create(&player)

	// 4. Get a token for the creator
	token, err := tests.GetTestUserToken(creator.ID, creator.Email)
	assert.NoError(t, err)

	// 5. Perform the delete request
	req, _ := http.NewRequest("DELETE", "/api/games/"+gameToCreate.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 6. Assert the results
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the game is deleted
	var deletedGame game.Game
	err = db.First(&deletedGame, "id = ?", gameToCreate.ID).Error
	assert.Error(t, err, "game should be deleted")

	// Check that the player is deleted
	var deletedPlayer game.Player
	err = db.First(&deletedPlayer, "game_id = ?", gameToCreate.ID).Error
	assert.Error(t, err, "player should be deleted")

	// Check that the save record is deleted
	var deletedSave game.Save
	err = db.First(&deletedSave, "game_id = ?", gameToCreate.ID).Error
	assert.Error(t, err, "save record should be deleted")

	// Check that the save file is deleted
	_, err = os.Stat(saveFilePath)
	assert.True(t, os.IsNotExist(err), "save file should be deleted")
}

func TestDeleteGame_NotCreator(t *testing.T) {
	db, router, err := tests.SetupTestEnvironment()
	assert.NoError(t, err)
	defer tests.TeardownTestEnvironment(db)

	// 1. Create a user and a game
	creator, err := tests.CreateTestUser(db, "creator@test.com")
	assert.NoError(t, err)
	notCreator, err := tests.CreateTestUser(db, "notcreator@test.com")
	assert.NoError(t, err)
	gameToCreate := game.Game{Name: "GameToDelete", CreatorID: creator.ID}
	db.Create(&gameToCreate)

	// 2. Get a token for the non-creator user
	token, err := tests.GetTestUserToken(notCreator.ID, notCreator.Email)
	assert.NoError(t, err)

	// 3. Perform the delete request
	req, _ := http.NewRequest("DELETE", "/api/games/"+gameToCreate.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 4. Assert the results
	assert.Equal(t, http.StatusForbidden, w.Code)
}
