package saves_test

import (
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"panzerstadt/async-multiplayer/tests/saves/helpers"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = helpers.SetupTestDB(db)
	require.NoError(t, err)

	return db
}

func TestCreateFakeUser(t *testing.T) {
	db := setupTestDB(t)

	user := helpers.CreateFakeUser(db)

	assert.NotEmpty(t, user.ID)
	assert.NotEmpty(t, user.Email)
	assert.Contains(t, []string{"google", "github", "discord"}, user.AuthProvider)

	// Verify user was created in database
	var dbUser helpers.User
	err := db.First(&dbUser, "id = ?", user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, user.Email, dbUser.Email)
}

func TestCreateFakeGame(t *testing.T) {
	db := setupTestDB(t)

	user := helpers.CreateFakeUser(db)
	game := helpers.CreateFakeGame(db, user.ID)

	assert.NotEmpty(t, game.ID)
	assert.NotEmpty(t, game.Name)
	assert.Equal(t, user.ID, game.CreatorID)

	// Verify game was created in database
	var dbGame helpers.Game
	err := db.First(&dbGame, "id = ?", game.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, game.Name, dbGame.Name)
}

func TestCreateFakeGameWithPlayers(t *testing.T) {
	db := setupTestDB(t)

	playerCount := 4
	game, users, players := helpers.CreateFakeGameWithPlayers(db, playerCount)

	assert.NotEmpty(t, game.ID)
	assert.Len(t, users, playerCount)
	assert.Len(t, players, playerCount)

	// Verify all players belong to the game
	for i, player := range players {
		assert.Equal(t, game.ID, player.GameID)
		assert.Equal(t, users[i].ID, player.UserID)
		assert.Equal(t, i, player.TurnOrder)
	}

	// Verify creator is the first player
	assert.Equal(t, game.CreatorID, users[0].ID)
	assert.Equal(t, 0, players[0].TurnOrder)
}

func TestCreateFakeSave(t *testing.T) {
	db := setupTestDB(t)

	game, users, _ := helpers.CreateFakeGameWithPlayers(db, 2)

	// Create a temporary file
	tempDir, err := helpers.CreateTestSaveDirectory()
	require.NoError(t, err)
	defer helpers.CleanupTestSaveDirectory(tempDir)

	filePath := filepath.Join(tempDir, "test_save.zip")
	err = helpers.CreateDummySaveFile(filePath, helpers.TypeZip, helpers.SizeMedium)
	require.NoError(t, err)

	save := helpers.CreateFakeSave(db, game.ID, users[0].ID, filePath)

	assert.NotEmpty(t, save.ID)
	assert.Equal(t, game.ID, save.GameID)
	assert.Equal(t, users[0].ID, save.UploadedBy)
	assert.Equal(t, filePath, save.FilePath)

	// Verify save was created in database
	var dbSave helpers.Save
	err = db.First(&dbSave, "id = ?", save.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, save.FilePath, dbSave.FilePath)
}

func TestCleanupTestData(t *testing.T) {
	db := setupTestDB(t)

	// Create test data
	game, users, players := helpers.CreateFakeGameWithPlayers(db, 3)

	tempDir, err := helpers.CreateTestSaveDirectory()
	require.NoError(t, err)
	defer helpers.CleanupTestSaveDirectory(tempDir)

	filePath := filepath.Join(tempDir, "test_save.sav")
	err = helpers.CreateDummySaveFile(filePath, helpers.TypeSav, helpers.SizeSmall)
	require.NoError(t, err)

	save := helpers.CreateFakeSave(db, game.ID, users[0].ID, filePath)

	// Verify data exists
	var count int64
	db.Model(&helpers.User{}).Count(&count)
	assert.Equal(t, int64(3), count)

	db.Model(&helpers.Game{}).Count(&count)
	assert.Equal(t, int64(1), count)

	db.Model(&helpers.Player{}).Count(&count)
	assert.Equal(t, int64(3), count)

	db.Model(&helpers.Save{}).Count(&count)
	assert.Equal(t, int64(1), count)

	// Clean up
	helpers.CleanupTestData(db)

	// Verify all data is gone
	db.Model(&helpers.User{}).Count(&count)
	assert.Equal(t, int64(0), count)

	db.Model(&helpers.Game{}).Count(&count)
	assert.Equal(t, int64(0), count)

	db.Model(&helpers.Player{}).Count(&count)
	assert.Equal(t, int64(0), count)

	db.Model(&helpers.Save{}).Count(&count)
	assert.Equal(t, int64(0), count)

	// Verify files are cleaned up
	assert.NotNil(t, save)    // Save object should still exist in memory
	assert.Len(t, players, 3) // Verify players were created
}

// Mock response structure for testing endpoint responses
type MockResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// Mock function to simulate GET /games/:id/saves/latest endpoint
// This will fail until the actual endpoint is implemented
//
// The real implementation should:
// 1. Validate that the game exists
// 2. Check if the user is authorized to access the game's saves (is a player)
// 3. Find the most recent save file for the game
// 4. Return the file with proper Content-Disposition header for download
func getLatestSave(gameID, userID uuid.UUID) MockResponse {
	// This is intentionally not implemented to make tests fail
	// The actual implementation will replace this
	return MockResponse{
		StatusCode: 500, // Intentionally wrong to make test fail
		Headers:    make(map[string]string),
		Body:       nil,
	}
}

// TestGetLatestSave tests the GET /games/:id/saves/latest endpoint
// This endpoint should allow players to download the most recent save file for a game
func TestGetLatestSave(t *testing.T) {
	db := setupTestDB(t)

	t.Run("successful case - 200 with correct Content-Disposition and binary body", func(t *testing.T) {
		game, users, _ := helpers.CreateFakeGameWithPlayers(db, 2)

		// Create a temporary file and save
		tempDir, err := helpers.CreateTestSaveDirectory()
		require.NoError(t, err)
		defer helpers.CleanupTestSaveDirectory(tempDir)

		filePath := filepath.Join(tempDir, "latest_save.zip")
		err = helpers.CreateDummySaveFile(filePath, helpers.TypeZip, helpers.SizeMedium)
		require.NoError(t, err)

		_ = helpers.CreateFakeSave(db, game.ID, users[0].ID, filePath)

		// Test the endpoint (this will fail until implemented)
		response := getLatestSave(game.ID, users[0].ID)

		// Expect 200 status code
		assert.Equal(t, 200, response.StatusCode, "Expected 200 status for successful save download")

		// Expect correct Content-Disposition header with filename
		contentDisposition, exists := response.Headers["Content-Disposition"]
		assert.True(t, exists, "Content-Disposition header should be present")
		assert.Contains(t, contentDisposition, "attachment", "Content-Disposition should indicate attachment")
		assert.Contains(t, contentDisposition, "filename=\"latest_save.zip\"", "Content-Disposition should include correct filename")

		// Expect binary body (non-empty)
		assert.NotNil(t, response.Body, "Response body should not be nil")
		assert.Greater(t, len(response.Body), 0, "Response body should contain binary data")
	})

	t.Run("no saves yet - 404", func(t *testing.T) {
		game, users, _ := helpers.CreateFakeGameWithPlayers(db, 2)

		// Don't create any saves for this game

		// Test the endpoint
		response := getLatestSave(game.ID, users[0].ID)

		// Expect 404 status code when no saves exist
		assert.Equal(t, 404, response.StatusCode, "Expected 404 status when no saves exist for the game")
	})

	t.Run("game not found - 404", func(t *testing.T) {
		// Use a non-existent game ID
		nonExistentGameID := uuid.New()
		userID := uuid.New()

		// Test the endpoint
		response := getLatestSave(nonExistentGameID, userID)

		// Expect 404 status code when game doesn't exist
		assert.Equal(t, 404, response.StatusCode, "Expected 404 status when game does not exist")
	})

	t.Run("unauthorized - 403", func(t *testing.T) {
		game, users, _ := helpers.CreateFakeGameWithPlayers(db, 2)

		// Create a save for this game
		tempDir, err := helpers.CreateTestSaveDirectory()
		require.NoError(t, err)
		defer helpers.CleanupTestSaveDirectory(tempDir)

		filePath := filepath.Join(tempDir, "unauthorized_save.zip")
		err = helpers.CreateDummySaveFile(filePath, helpers.TypeZip, helpers.SizeMedium)
		require.NoError(t, err)

		_ = helpers.CreateFakeSave(db, game.ID, users[0].ID, filePath)

		// Use a user ID that is not a player in this game
		unauthorizedUserID := uuid.New()

		// Test the endpoint
		response := getLatestSave(game.ID, unauthorizedUserID)

		// Expect 403 status code when user is not authorized to access the game's saves
		assert.Equal(t, 403, response.StatusCode, "Expected 403 status when user is not authorized to access game saves")
	})
}
