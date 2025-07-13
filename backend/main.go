package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"panzerstadt/async-multiplayer/game"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize Gin router with custom error middleware
	r := gin.Default()
	r.Use(game.ErrorHandlingMiddleware())

	// Initialize SQLite database using GORM
	db, err := gorm.Open(sqlite.Open("game.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Perform initial database migration
	db.AutoMigrate(&game.User{}, &game.Game{}, &game.Player{}, &game.Save{})

	// Initialize OAuth
	game.InitOAuth()

	// Define API routes
	r.POST("/create-game", game.CreateGameHandler(db))
	r.POST("/join-game/:id", game.JoinGameHandler(db))
	r.GET("/auth/google/login", game.GoogleLoginHandler)
	r.GET("/auth/google/callback", game.GoogleCallbackHandler(db))

	// Authenticated routes
	authed := r.Group("/api")
	authed.Use(game.AuthMiddleware())
	authed.GET("/user/games", game.GetUserGamesHandler(db))
	r.GET("/games/:id", game.GetGameHandler(db))

	// Create rate limited upload endpoint
	savesGroup := r.Group("/games/:id/saves")
	savesGroup.Use(game.RateLimitMiddleware(10, time.Minute)) // 10 requests per minute
	savesGroup.POST("", game.UploadSaveHandler(db))
	savesGroup.GET("/latest", game.GetLatestSaveHandler(db))

	// Start the server
	r.Run() // listens and serves on 0.0.0.0:8080 by default
}
