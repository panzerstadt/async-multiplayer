package game

import (
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Middleware for centralized error handling
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Retrieve error if it exists
		errors := c.Errors.ByType(gin.ErrorTypePrivate)
		if len(errors) > 0 {
			// Map errors to status codes
			switch errors[0].Err {
			case gorm.ErrRecordNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
			case gorm.ErrInvalidData:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
			default:
				if os.IsNotExist(errors[0].Err) {
					c.JSON(http.StatusGone, gin.H{"error": "resource removed"})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				}
			}
		}
	}
}

// Rate limiting middleware
func RateLimitMiddleware(maxRequests int, duration time.Duration) gin.HandlerFunc {
	var rateLimiter = make(chan time.Time, maxRequests)
	go func() {
		ticker := time.NewTicker(duration / time.Duration(maxRequests))
		defer ticker.Stop()
		for t := range ticker.C {
			rateLimiter <- t
		}
	}()

	return func(c *gin.Context) {
		select {
		case <-rateLimiter:
			c.Next()
		default:
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
		}
	}
}

// Determine MIME type of file buffer by content with a default assumption
func detectMimeType(file io.Reader) (string, error) {
	// Read up to 512 bytes to detect mime type
	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil && err != io.EOF {
		return "", err
	}

	fileType := http.DetectContentType(buf)
	return fileType, nil
}

func GetLatestSaveHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("User-ID")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		gameIDStr := c.Param("id")
		gameID, err := uuid.Parse(gameIDStr)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}

		// Check if game exists
		var game Game
		if err := db.First(&game, "id = ?", gameID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}

		// Check if user is a member of the game
		var player Player
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}
		if err := db.Where("user_id = ? AND game_id = ?", userUUID, gameID).First(&player).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this game"})
			return
		}

		var save Save
		if err := db.Where("game_id = ? ", gameID).Order("created_at DESC").First(&save).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no saves found"})
			return
		}

		filePath := save.FilePath
		file, err := os.Open(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusGone, gin.H{"error": "save file has been removed"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open save file"})
			return
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get file info"})
			return
		}

		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_latest.sav\"", gameID))
		c.Writer.Header().Set("Last-Modified", fileInfo.ModTime().UTC().Format(http.TimeFormat))
		c.Writer.Header().Set("ETag", fmt.Sprintf("%x-%x", fileInfo.ModTime().Unix(), fileInfo.Size()))

		// Copy the file to the response writer
		if _, err := io.Copy(c.Writer, file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stream file"})
			return
		}
	}
}

// Security utility functions
func sanitizeFilename(filename string) string {
	// Remove path separators and dangerous characters
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.TrimSpace(filename)

	if filename == "" || filename == "." {
		return ""
	}

	return filename
}

// isValidFileExtension checks if the file extension is allowed using constant-time comparison
func IsValidFileExtension(filename string) bool {
	validExtensions := []string{".zip", ".sav"}
	lowerFilename := strings.ToLower(filename)

	for _, ext := range validExtensions {
		if strings.HasSuffix(lowerFilename, ext) {
			// Use constant-time comparison for security
			if subtle.ConstantTimeCompare([]byte(ext), []byte(lowerFilename[len(lowerFilename)-len(ext):])) == 1 {
				return true
			}
		}
	}
	return false
}

// isPathSafe checks for path traversal attacks
func isPathSafe(filePath, baseDir string) bool {
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}

	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}

	return strings.HasPrefix(absFilePath, absBaseDir)
}

// advanceTurn handles turn management: mark current turn complete and assign next player
func advanceTurn(db *gorm.DB, gameID uuid.UUID) error {
	// Get all players in the game, ordered by turn order
	var players []Player
	if err := db.Where("game_id = ?", gameID).Order("turn_order ASC").Find(&players).Error; err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	if len(players) == 0 {
		return fmt.Errorf("no players found for game")
	}

	// Get current game state
	var game Game
	if err := db.First(&game, "id = ?", gameID).Error; err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Find next player in turn order
	var nextPlayerID *uuid.UUID
	if game.CurrentTurnID == nil {
		// First turn, assign to first player
		nextPlayerID = &players[0].ID
	} else {
		// Find current player index
		currentPlayerIndex := -1
		for i, player := range players {
			if player.ID == *game.CurrentTurnID {
				currentPlayerIndex = i
				break
			}
		}

		if currentPlayerIndex == -1 {
			// Current player not found, reset to first player
			nextPlayerID = &players[0].ID
		} else {
			// Move to next player (wrap around if at end)
			nextIndex := (currentPlayerIndex + 1) % len(players)
			nextPlayerID = &players[nextIndex].ID
		}
	}

	// Update game with next player's turn
	if err := db.Model(&game).Update("current_turn_id", nextPlayerID).Error; err != nil {
		return fmt.Errorf("failed to update current turn: %w", err)
	}

	return nil
}

type CreateGameRequest struct {
	Name string `json:"name" binding:"required"`
}

func CreateGameHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateGameRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}

		// Check if game name already exists
		var existingGame Game
		if err := db.Where("name = ?", req.Name).First(&existingGame).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name already exists"})
			return
		}

		// Create new game
		game := Game{
			Name:      req.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := db.Create(&game).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Game created", "game_id": game.ID})
	}
}

func JoinGameHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for authentication (placeholder for now)
		userID := c.GetHeader("User-ID")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		gameIDStr := c.Param("id")
		gameID, err := uuid.Parse(gameIDStr)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}

		// Check if game exists
		var game Game
		if err := db.First(&game, "id = ?", gameID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}

		// Parse user ID
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		// Check if user is already in the game
		var existingPlayer Player
		if err := db.Where("user_id = ? AND game_id = ?", userUUID, gameID).First(&existingPlayer).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "already a participant"})
			return
		}

		// Add user to game
		var maxTurnOrder int
		db.Model(&Player{}).Where("game_id = ?", gameID).Select("COALESCE(MAX(turn_order), -1)").Scan(&maxTurnOrder)

		player := Player{
			UserID:    userUUID,
			GameID:    gameID,
			TurnOrder: maxTurnOrder + 1,
		}

		if err := db.Create(&player).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join game"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Joined game", "player_id": player.ID})
	}
}

func GetGameHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		gameIDStr := c.Param("id")
		gameID, err := uuid.Parse(gameIDStr)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}

		var game Game
		if err := db.Preload("Players").First(&game, "id = ?", gameID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}

		c.JSON(http.StatusOK, game)
	}
}

func UploadSaveHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Auth & membership check
		userID := c.GetHeader("User-ID")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		gameIDStr := c.Param("id")
		gameID, err := uuid.Parse(gameIDStr)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}

		// Parse user ID
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		// Check if game exists
		var game Game
		if err := db.First(&game, "id = ?", gameID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}

		// Check if user is a member of the game
		var player Player
		if err := db.Where("user_id = ? AND game_id = ?", userUUID, gameID).First(&player).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this game"})
			return
		}

		// 2. Accept multipart upload with disk buffer limits
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file upload failed"})
			return
		}
		defer file.Close()

		// MIME sniffing
		mimeType, err := detectMimeType(file)
		if err != nil || !strings.HasPrefix(mimeType, "application/zip") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type"})
			return
		}

		// Validate file type/size
		const maxFileSize = 100 * 1024 * 1024 // 100MB limit
		if header.Size > maxFileSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large"})
			return
		}

		// Sanitize and validate filename
		filename := sanitizeFilename(header.Filename)
		if filename == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}

		// Validate file extension (constant-time check)
		if !IsValidFileExtension(filename) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type"})
			return
		}

		// 4. Save via fileStorage service
		saveDir := fmt.Sprintf("saves/%s", gameID)
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create save directory"})
			return
		}

		// Create unique filename to prevent conflicts
		uniqueFilename := fmt.Sprintf("%s_%s", uuid.New().String(), filename)
		filePath := filepath.Join(saveDir, uniqueFilename)

		// Check for path traversal
		if !isPathSafe(filepath.Clean(filePath), saveDir) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file path"})
			return
		}

		// Create the file
		outFile, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create file"})
			return
		}
		defer outFile.Close()

		// Copy uploaded file to destination
		if _, err := io.Copy(outFile, file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
			return
		}

		// Record row in game_saves table
		save := Save{
			GameID:     gameID,
			FilePath:   filePath,
			UploadedBy: userUUID,
			CreatedAt:  time.Now(),
		}

		if err := db.Create(&save).Error; err != nil {
			// Clean up the file if database insert fails
			os.Remove(filePath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file record"})
			return
		}

		// 5. Invoke turn-manager: mark current turn complete & assign next player
        if err := advanceTurn(db, gameID); err != nil {
            // Log error but don't fail the request
            fmt.Printf("Warning: failed to advance turn for game %s: %v\n", gameID, err)
        }

        // Notify other players
        var players []Player
        db.Where("game_id = ? AND user_id != ?", gameID, userUUID).Find(&players)
        for _, p := range players {
            // In a real app, you'd get the user's email from the User model
            // For now, we'll use a placeholder
            _ = (&EmailNotifier{}).Notify(p.UserID.String(), fmt.Sprintf("New save uploaded for game %s!", game.Name))
        }

        // 6. Respond 201 with save metadata
        c.JSON(http.StatusCreated, gin.H{
            "message":     "save uploaded successfully",
            "save_id":     save.ID,
            "game_id":     save.GameID,
            "file_path":   save.FilePath,
            "uploaded_by": save.UploadedBy,
            "created_at":  save.CreatedAt,
        })
    }
}
