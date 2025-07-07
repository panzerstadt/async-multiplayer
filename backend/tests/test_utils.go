package tests

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"panzerstadt/async-multiplayer/game"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	db.AutoMigrate(&game.User{}, &game.Game{}, &game.Player{}, &game.Save{})
	return db
}

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.POST("/create-game", game.CreateGameHandler(db))
	r.POST("/join-game/:id", game.JoinGameHandler(db))
	r.GET("/games/:id", game.GetGameHandler(db))
	r.GET("/auth/google/login", game.GoogleLoginHandler)
	r.GET("/auth/google/callback", game.GoogleCallbackHandler(db))
	
	// Group save-related routes
	savesGroup := r.Group("/games/:id/saves")
	savesGroup.POST("", game.UploadSaveHandler(db))
	savesGroup.GET("/latest", game.GetLatestSaveHandler(db))
	return r
}
