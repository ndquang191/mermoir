package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"memoir/api/db"
	"memoir/api/handlers"
	"memoir/api/queue"
)

func main() {
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		log.Fatal("DB_DSN env var is required")
	}

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		log.Fatal("RABBITMQ_URL env var is required")
	}

	pool, err := db.Connect(dbDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := db.InitSchema(pool); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	pub, err := queue.NewPublisher(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer pub.Close()

	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "/storage"
	}

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	entryHandler := handlers.NewEntryHandler(pool)
	uploadHandler := handlers.NewUploadHandler(pool, pub, storagePath)

	api := r.Group("/api")
	{
		api.GET("/entries", entryHandler.GetEntries)
		api.POST("/entries", entryHandler.CreateEntry)
		api.POST("/entries/:id/photos", uploadHandler.UploadPhoto)
	}

	log.Println("API server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
