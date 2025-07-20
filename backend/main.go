package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"panzerstadt/async-multiplayer/config"
	"panzerstadt/async-multiplayer/game"
	"panzerstadt/async-multiplayer/sse"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// Initialize Gin router with custom error middleware
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(game.ErrorHandlingMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendUrl},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize SSE Manager
	sseManager := sse.NewSSEManager()
	go sseManager.Run()

	// Initialize SQLite database using GORM
	db, err := gorm.Open(sqlite.Open("game.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Perform initial database migration
	db.AutoMigrate(&game.User{}, &game.Game{}, &game.Player{}, &game.Save{})

	// Initialize OAuth
	game.InitOAuth(cfg)

	// Initialize Mailgun Notifier
	mailgunNotifier := game.NewMailgunNotifier(cfg)

	// Define API routes
	r.POST("/create-game", game.AuthMiddleware(cfg), game.CreateGameHandler(db))
	r.POST("/join-game/:id", game.JoinGameHandler(db))
	r.GET("/auth/google/login", game.GoogleLoginHandler(cfg))
	r.GET("/auth/google/callback", game.GoogleCallbackHandler(db, cfg))

	// SSE endpoint
	r.GET("/sse/notifications", func(c *gin.Context) {
		// TODO: serve SSE per game room
		// 1. get user, get gameID (from params)
		// 2. serve only to the same game
		sse.ServeSSE(sseManager, c)
	})

	// Authenticated routes
	authed := r.Group("/api")
	authed.Use(game.AuthMiddleware(cfg))
	authed.GET("/user/games", game.GetUserGamesHandler(db))
	authed.DELETE("/games/:id", game.DeleteGameHandler(db))
	r.GET("/games/:id", game.GetGameHandler(db))

	// Create rate limited upload endpoint
	savesGroup := r.Group("/games/:id/saves")
	savesGroup.Use(game.AuthMiddleware(cfg))
	savesGroup.Use(game.RateLimitMiddleware(10, time.Minute)) // 10 requests per minute
	savesGroup.POST("", game.UploadSaveHandler(db, sseManager, mailgunNotifier))

	savesGroup.GET("/latest", game.GetLatestSaveHandler(db))

	msgGroup := r.Group("games/:id/broadcast")
	msgGroup.Use(game.AuthMiddleware(cfg))
	msgGroup.Use(game.RateLimitMiddleware(100, time.Minute)) // 10 requests per minute
	msgGroup.POST("", game.MessageHandler(db, sseManager))

	// Start the server
	r.Run() // listens and serves on 0.0.0.0:8080 by default
}
