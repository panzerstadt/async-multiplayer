package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRouter() (*gin.Engine, *gorm.DB) {
	r := gin.Default()
	// Initialize SQLite database using GORM
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{}, &Game{}, &Player{}, &Save{})

	r.POST("/create-game", createGameHandler(db))
	r.POST("/join-game/:id", joinGameHandler(db))

	return r, db
}

func TestCreateGame_Success(t *testing.T) {
	r, _ := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/create-game", bytes.NewBufferString(`{"name":"Game 1"}`))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Game created")
}

func TestCreateGame_NameExists(t *testing.T) {
	r, db := setupRouter()

	// Creating initial game with name
	db.Create(&Game{Name: "Game 1"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/create-game", bytes.NewBufferString(`{"name":"Game 1"}`))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "name already exists")
}

func TestCreateGame_MissingName(t *testing.T) {
	r, _ := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/create-game", bytes.NewBufferString(`{}`))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "name is required")
}

func TestJoinGame_Success(t *testing.T) {
	r, db := setupRouter()

	// Creating initial game for joining
	game := Game{Name: "Game to Join"}
	db.Create(&game)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/join-game/"+game.ID.String(), nil)
	req.Header.Set("User-ID", "550e8400-e29b-41d4-a716-446655440000")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Joined game")
}

func TestJoinGame_NotFound(t *testing.T) {
	r, _ := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/join-game/invalid-id", nil)
	req.Header.Set("User-ID", "550e8400-e29b-41d4-a716-446655440000")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "game not found")
}

func TestJoinGame_AlreadyInGame(t *testing.T) {
	r, db := setupRouter()

	// Creating initial game and player
	user := User{Email: "test@example.com"}
	game := Game{Name: "Game to Join"}
	db.Create(&game)
	db.Create(&user)
	db.Create(&Player{UserID: user.ID, GameID: game.ID})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/join-game/"+game.ID.String(), nil)
	req.Header.Set("User-ID", user.ID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "already a participant")
}

func TestJoinGame_MissingAuthentication(t *testing.T) {
	r, _ := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/join-game/some-id", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authentication required")
}

// Tests for POST /games/:id/saves endpoint

func TestUploadSave_Success(t *testing.T) {
	r, db := setupRouter()

	// Create user, game, and player
	user := User{Email: "test@example.com"}
	game := Game{Name: "Test Game"}
	db.Create(&user)
	db.Create(&game)
	db.Create(&Player{UserID: user.ID, GameID: game.ID, TurnOrder: 1})

	// Create multipart form data with a valid save file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("save", "test.sav")
	part.Write([]byte("mock save file content"))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/games/"+game.ID.String()+"/saves", body)
	req.Header.Set("User-ID", user.ID.String())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "save_id")
	assert.Contains(t, w.Body.String(), "file_path")
	assert.Contains(t, w.Body.String(), "uploaded_by")
	assert.Contains(t, w.Body.String(), "created_at")
}

func TestUploadSave_GameNotFound(t *testing.T) {
	r, db := setupRouter()

	// Create user but no game
	user := User{Email: "test@example.com"}
	db.Create(&user)

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("save", "test.sav")
	part.Write([]byte("mock save file content"))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/games/550e8400-e29b-41d4-a716-446655440000/saves", body)
	req.Header.Set("User-ID", user.ID.String())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "game not found")
}

func TestUploadSave_UserNotInGame(t *testing.T) {
	r, db := setupRouter()

	// Create user and game, but user is not a player in the game
	user := User{Email: "test@example.com"}
	game := Game{Name: "Test Game"}
	db.Create(&user)
	db.Create(&game)
	// Note: Not creating a Player record for this user in this game

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("save", "test.sav")
	part.Write([]byte("mock save file content"))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/games/"+game.ID.String()+"/saves", body)
	req.Header.Set("User-ID", user.ID.String())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "user not in game")
}

func TestUploadSave_InvalidMimeType(t *testing.T) {
	r, db := setupRouter()

	// Create user, game, and player
	user := User{Email: "test@example.com"}
	game := Game{Name: "Test Game"}
	db.Create(&user)
	db.Create(&game)
	db.Create(&Player{UserID: user.ID, GameID: game.ID, TurnOrder: 1})

	// Create multipart form data with invalid file type
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("save", "test.txt")
	part.Write([]byte("this is not a save file"))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/games/"+game.ID.String()+"/saves", body)
	req.Header.Set("User-ID", user.ID.String())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	assert.Contains(t, w.Body.String(), "invalid file type")
}

func TestUploadSave_FileSizeExceedsLimit(t *testing.T) {
	r, db := setupRouter()

	// Create user, game, and player
	user := User{Email: "test@example.com"}
	game := Game{Name: "Test Game"}
	db.Create(&user)
	db.Create(&game)
	db.Create(&Player{UserID: user.ID, GameID: game.ID, TurnOrder: 1})

	// Create multipart form data with oversized file (simulate large file)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("save", "huge.sav")
	// Create a large buffer to simulate file size exceeding limit
	largeContent := make([]byte, 11*1024*1024) // 11MB, assuming 10MB limit
	part.Write(largeContent)
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/games/"+game.ID.String()+"/saves", body)
	req.Header.Set("User-ID", user.ID.String())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Contains(t, w.Body.String(), "file size exceeds maximum")
}

func TestUploadSave_MissingAuthentication(t *testing.T) {
	r, db := setupRouter()

	// Create game
	game := Game{Name: "Test Game"}
	db.Create(&game)

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("save", "test.sav")
	part.Write([]byte("mock save file content"))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/games/"+game.ID.String()+"/saves", body)
	// Note: No User-ID header set
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authentication required")
}

func TestUploadSave_TurnAdvancementCallback(t *testing.T) {
	r, db := setupRouter()

	// Create users, game, and players for turn-based game
	user1 := User{Email: "player1@example.com"}
	user2 := User{Email: "player2@example.com"}
	game := Game{Name: "Turn-based Game"}
	db.Create(&user1)
	db.Create(&user2)
	db.Create(&game)
	db.Create(&Player{UserID: user1.ID, GameID: game.ID, TurnOrder: 1})
	db.Create(&Player{UserID: user2.ID, GameID: game.ID, TurnOrder: 2})

	// Set current turn to user1
	game.CurrentTurnID = &user1.ID
	db.Save(&game)

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("save", "turn1.sav")
	part.Write([]byte("player 1 turn save"))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/games/"+game.ID.String()+"/saves", body)
	req.Header.Set("User-ID", user1.ID.String())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify turn was advanced to next player
	var updatedGame Game
	db.First(&updatedGame, "id = ?", game.ID)
	assert.Equal(t, user2.ID, *updatedGame.CurrentTurnID)
}

func TestUploadSave_InvalidGameID(t *testing.T) {
	r, db := setupRouter()

	// Create user
	user := User{Email: "test@example.com"}
	db.Create(&user)

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("save", "test.sav")
	part.Write([]byte("mock save file content"))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/games/invalid-uuid/saves", body)
	req.Header.Set("User-ID", user.ID.String())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "game not found")
}
