package main

import (
	"log"

	"quivra-backend/config"
	"quivra-backend/database"
	"quivra-backend/handlers"
	"quivra-backend/services"
	"quivra-backend/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// 設定を読み込み
	cfg := config.LoadConfig()

	// データベース接続
	db, err := database.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// データベースはDockerコンテナの初期化時に自動でセットアップされます

	// サービスを初期化
	roomService := services.NewRoomService(db)
	questionService := services.NewQuestionService(db)
	gameService := services.NewGameService(db)
	buzzManager := services.NewBuzzManager()

	// WebSocket Hubを初期化
	hub := websocket.NewHub()
	go hub.Run()

	// WebSocketハンドラーを初期化
	wsHandler := websocket.NewWSHandler(hub, roomService, questionService, gameService, buzzManager)

	// HTTPハンドラーを初期化
	roomHandler := handlers.NewRoomHandler(roomService)
	questionHandler := handlers.NewQuestionHandler(questionService)

	// Ginルーターを設定
	router := gin.Default()

	// CORS設定
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API ルート
	api := router.Group("/api")
	{
		// ルーム関連
		api.POST("/rooms", roomHandler.CreateRoom)
		api.GET("/rooms/:roomId", roomHandler.GetRoom)
		api.POST("/rooms/join", roomHandler.JoinRoom)

		// 問題関連
		api.POST("/questions", questionHandler.CreateQuestion)
		api.GET("/questions", questionHandler.GetQuestions)
		api.GET("/questions/:id", questionHandler.GetQuestion)
	}

	// WebSocket エンドポイント
	router.GET("/ws", wsHandler.HandleWebSocket)

	// サーバーを起動
	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
