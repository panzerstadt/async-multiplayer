package tests

import (
	"io/ioutil"
	"os"
	"panzerstadt/async-multiplayer/config"
	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/sse"
	"path/filepath"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockNotifier is a mock implementation of the Notifier interface for testing.
type MockNotifier struct {
	// LastRecipientEmail holds the email of the last recipient.
	LastRecipientEmail string
	// LastSubject holds the subject of the last notification.
	LastSubject string
	// LastBody holds the body of the last notification.
	LastBody string
	// Err is the error to return from Notify.
	Err error
}

// NewMockNotifier creates a new MockNotifier.
func NewMockNotifier() *MockNotifier {
	return &MockNotifier{}
}

// Notify captures the notification details and returns a predefined error, if any.
func (m *MockNotifier) Notify(recipientEmail string, subject string, body string) error {
	m.LastRecipientEmail = recipientEmail
	m.LastSubject = subject
	m.LastBody = body
	return m.Err
}

// MockSSEManager is a mock implementation of the sse.Broadcaster interface.
type MockSSEManager struct{}

// BroadcastMessage is a no-op for the mock manager.
func (m *MockSSEManager) BroadcastMessage(eventType string, data interface{}) {}

// AddClient is a no-op for the mock manager.
func (m *MockSSEManager) AddClient(client chan string) {}

// RemoveClient is a no-op for the mock manager.
func (m *MockSSEManager) RemoveClient(client chan string) {}

// Run is a no-op for the mock manager.
func (m *MockSSEManager) Run() {}

var jwtKey = []byte("test_secret_key")

func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	db.AutoMigrate(&game.User{}, &game.Game{}, &game.Player{}, &game.Save{})
	return db
}

func SetupRouter(db *gorm.DB, cfg config.Config, notifier game.Notifier) *gin.Engine {
	r := gin.Default()
	sseManager := sse.NewSSEManager()
	r.POST("/create-game", game.CreateGameHandler(db))
	r.POST("/join-game/:id", game.JoinGameHandler(db))
	r.GET("/games/:id", game.GetGameHandler(db))
	r.GET("/auth/google/login", game.GoogleLoginHandler)
	r.GET("/auth/google/callback", game.GoogleCallbackHandler(db, cfg))

	// Group save-related routes
	savesGroup := r.Group("/games/:id/saves")
	savesGroup.POST("", game.UploadSaveHandler(db, sseManager, notifier))
	savesGroup.GET("/latest", game.GetLatestSaveHandler(db))
	return r
}

// SetupTestEnvironment initializes the database and router for testing.
func SetupTestEnvironment() (*gorm.DB, *gin.Engine, config.Config, error) {
	// Create a temporary directory for config
	tempDir, err := ioutil.TempDir("", "test_config")
	if err != nil {
		return nil, nil, config.Config{}, err
	}
	// Defer cleanup of the temporary directory
	defer os.RemoveAll(tempDir)

	// Copy config.yaml to the temporary directory
	sourcePath := "../../config.yaml" // Path relative to backend/tests/
	destPath := filepath.Join(tempDir, "config.yaml")

	input, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return nil, nil, config.Config{}, err
	}

	err = ioutil.WriteFile(destPath, input, 0644)
	if err != nil {
		return nil, nil, config.Config{}, err
	}

	// Load configuration from the temporary directory
	cfg, err := config.LoadConfig(tempDir)
	if err != nil {
		return nil, nil, config.Config{}, err
	}

	// Set up the in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, nil, config.Config{}, err
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&game.User{}, &game.Game{}, &game.Player{}, &game.Save{}); err != nil {
		return nil, nil, config.Config{}, err
	}

	// Set up the Gin router
	r := gin.Default()
	sseManager := &MockSSEManager{}

	// Public routes
	r.POST("/create-game", game.AuthMiddleware(cfg), game.CreateGameHandler(db))
	r.POST("/join-game/:id", game.AuthMiddleware(cfg), game.JoinGameHandler(db))
	r.GET("/games/:id", game.GetGameHandler(db))
	r.GET("/auth/google/login", game.GoogleLoginHandler)
	r.GET("/auth/google/callback", game.GoogleCallbackHandler(db, cfg))

	// Authenticated routes
	authed := r.Group("/api")
	authed.Use(game.AuthMiddleware(cfg))
	authed.GET("/user/games", game.GetUserGamesHandler(db))
	authed.DELETE("/games/:id", game.DeleteGameHandler(db))

	// Group save-related routes
	savesGroup := r.Group("/games/:id/saves")
	savesGroup.Use(game.AuthMiddleware(cfg))
	// In a real test setup, you would pass a mock notifier here.
	// For now, we'll use the actual notifier but this setup allows for mocking.
	notifier := game.NewMailgunNotifier(cfg)
	savesGroup.POST("", game.UploadSaveHandler(db, sseManager, notifier))
	savesGroup.GET("/latest", game.GetLatestSaveHandler(db))

	return db, r, cfg, nil
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
func GetTestUserToken(userID uuid.UUID, email string, cfg config.Config) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"exp":   expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JwtSecret))
}

// SetupTestEnvironmentWithNotifier initializes the database and router for testing with a custom notifier.
func SetupTestEnvironmentWithNotifier(t *testing.T, notifier game.Notifier) (*gorm.DB, *gin.Engine, config.Config) {
	// Create a temporary directory for config
	tempDir, err := ioutil.TempDir("", "test_config")
	require.NoError(t, err)
	// Defer cleanup of the temporary directory
	defer os.RemoveAll(tempDir)

	// Copy config.yaml to the temporary directory
	sourcePath := "../../config.yaml" // Path relative to backend/tests/
	destPath := filepath.Join(tempDir, "config.yaml")

	input, err := ioutil.ReadFile(sourcePath)
	require.NoError(t, err)

	err = ioutil.WriteFile(destPath, input, 0644)
	require.NoError(t, err)

	// Load configuration from the temporary directory
	cfg, err := config.LoadConfig(tempDir)
	require.NoError(t, err)

	// Set up the in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&game.User{}, &game.Game{}, &game.Player{}, &game.Save{})
	require.NoError(t, err)

	// Set up the Gin router
	r := gin.Default()
	sseManager := &MockSSEManager{}

	// Public routes
	r.POST("/create-game", game.AuthMiddleware(cfg), game.CreateGameHandler(db))
	r.POST("/join-game/:id", game.AuthMiddleware(cfg), game.JoinGameHandler(db))
	r.GET("/games/:id", game.GetGameHandler(db))
	r.GET("/auth/google/login", game.GoogleLoginHandler)
	r.GET("/auth/google/callback", game.GoogleCallbackHandler(db, cfg))

	// Authenticated routes
	authed := r.Group("/api")
	authed.Use(game.AuthMiddleware(cfg))
	authed.GET("/user/games", game.GetUserGamesHandler(db))
	authed.DELETE("/games/:id", game.DeleteGameHandler(db))

	// Group save-related routes
	savesGroup := r.Group("/games/:id/saves")
	savesGroup.Use(game.AuthMiddleware(cfg))
	savesGroup.POST("", game.UploadSaveHandler(db, sseManager, notifier))
	savesGroup.GET("/latest", game.GetLatestSaveHandler(db))

	return db, r, cfg
}
