package tests

import (
	"os"
	"panzerstadt/async-multiplayer/game"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	socketio "github.com/googollee/go-socket.io"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var jwtKey = []byte("test_secret_key")

func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	db.AutoMigrate(&game.User{}, &game.Game{}, &game.Player{}, &game.Save{})
	return db
}

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	server := socketio.NewServer(nil)
	r.POST("/create-game", game.CreateGameHandler(db))
	r.POST("/join-game/:id", game.JoinGameHandler(db))
	r.GET("/games/:id", game.GetGameHandler(db))
	r.GET("/auth/google/login", game.GoogleLoginHandler)
	r.GET("/auth/google/callback", game.GoogleCallbackHandler(db))

	// Group save-related routes
	savesGroup := r.Group("/games/:id/saves")
	savesGroup.POST("", game.UploadSaveHandler(db, server))
	savesGroup.GET("/latest", game.GetLatestSaveHandler(db))
	return r
}

// SetupTestEnvironment initializes the database and router for testing.
func SetupTestEnvironment() (*gorm.DB, *gin.Engine, error) {
	// Set up the in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&game.User{}, &game.Game{}, &game.Player{}, &game.Save{}); err != nil {
		return nil, nil, err
	}

	// Set up the Gin router
	r := gin.Default()
	server := socketio.NewServer(nil)

	// Public routes
	r.POST("/create-game", game.AuthMiddleware(), game.CreateGameHandler(db))
	r.POST("/join-game/:id", game.AuthMiddleware(), game.JoinGameHandler(db))
	r.GET("/games/:id", game.GetGameHandler(db))
	r.GET("/auth/google/login", game.GoogleLoginHandler)
	r.GET("/auth/google/callback", game.GoogleCallbackHandler(db))

	// Authenticated routes
	authed := r.Group("/api")
	authed.Use(game.AuthMiddleware())
	authed.GET("/user/games", game.GetUserGamesHandler(db))
	authed.DELETE("/games/:id", game.DeleteGameHandler(db))

	// Group save-related routes
	savesGroup := r.Group("/games/:id/saves")
	savesGroup.Use(game.AuthMiddleware())
	savesGroup.POST("", game.UploadSaveHandler(db, server))
	savesGroup.GET("/latest", game.GetLatestSaveHandler(db))

	return db, r, nil
}

// TeardownTestEnvironment closes the database connection.
func TeardownTestEnvironment(db *gorm.DB) {
	sqlDB, _ := db.DB()
	sqlDB.Close()
}

// CreateTestUser creates a user for testing purposes.
func CreateTestUser(db *gorm.DB, email string) (*game.User, error) {
	user := &game.User{
		Email:        email,
		AuthProvider: "google",
	}
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetTestUserToken generates a JWT token for a test user.
func GetTestUserToken(userID uuid.UUID, email string) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"exp":   expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
