package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
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

	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "/storage"
	}

	numWorkers := 4
	if v := os.Getenv("NUM_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			numWorkers = n
		}
	}

	pool, err := pgxpool.New(context.Background(), dbDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan ImageJob, numWorkers*2)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		go runWorker(ctx, i, jobs, pool, storagePath)
	}

	// Start consumer
	go func() {
		if err := Consume(ctx, conn, jobs); err != nil {
			log.Printf("consumer error: %v", err)
			cancel()
		}
	}()

	log.Printf("Worker started with %d goroutines", numWorkers)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker...")
	cancel()
}
