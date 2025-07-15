package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"panzerstadt/async-multiplayer/config"
	"panzerstadt/async-multiplayer/game"
)

func main() {
	// Load configuration
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// Initialize Gin router with custom error middleware
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(game.ErrorHandlingMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{config.FrontendUrl},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize Socket.IO server
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		return nil
	})

	go server.Serve()
	defer server.Close()

	r.GET("/socket.io/*any", gin.WrapH(server))
	r.POST("/socket.io/*any", gin.WrapH(server))

	// Initialize SQLite database using GORM
	db, err := gorm.Open(sqlite.Open("game.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Perform initial database migration
	db.AutoMigrate(&game.User{}, &game.Game{}, &game.Player{}, &game.Save{})

	// Initialize OAuth
	game.InitOAuth(config)

	// Define API routes
	r.POST("/create-game", game.AuthMiddleware(config), game.CreateGameHandler(db))
	r.POST("/join-game/:id", game.JoinGameHandler(db))
	r.GET("/auth/google/login", game.GoogleLoginHandler)
	r.GET("/auth/google/callback", game.GoogleCallbackHandler(db, config))

	// Authenticated routes
	authed := r.Group("/api")
	authed.Use(game.AuthMiddleware(config))
	authed.GET("/user/games", game.GetUserGamesHandler(db))
	authed.DELETE("/games/:id", game.DeleteGameHandler(db))
	r.GET("/games/:id", game.GetGameHandler(db))

	// Create rate limited upload endpoint
	savesGroup := r.Group("/games/:id/saves")
	savesGroup.Use(game.AuthMiddleware(config))
	savesGroup.Use(game.RateLimitMiddleware(10, time.Minute)) // 10 requests per minute
	savesGroup.POST("", game.UploadSaveHandler(db, server))
	savesGroup.GET("/latest", game.GetLatestSaveHandler(db))

	// Start the server
	r.Run() // listens and serves on 0.0.0.0:8080 by default
}
