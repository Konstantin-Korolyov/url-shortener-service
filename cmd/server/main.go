package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Konstantin-Korolyov/url-shortener-go/internal/cache"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/database"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/handlers"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/kafka"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/repository"
)

func main() {
	// PostgreSQL
	dbCfg := database.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "admin",
		Password: "securepassword123",
		DBName:   "shortener_db",
	}
	dbPool, err := database.NewPool(dbCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	// Redis
	redisClient, err := cache.NewRedisClient("localhost:6379", "redispassword123", 0)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Kafka producer
	producer := kafka.NewProducer([]string{"localhost:9093"}, "clicks")
	defer producer.Close()

	// Репозиторий
	urlRepo := repository.NewURLRepository(dbPool)

	// Обработчики с продюсером
	urlHandlers := handlers.NewURLHandlers(urlRepo, redisClient, producer)

	// Kafka consumer (запускаем в фоне)
	consumerCfg := kafka.ConsumerConfig{
		Brokers: []string{"localhost:9093"},
		Topic:   "clicks",
		GroupID: "click-consumers",
	}
	consumer := kafka.NewClickEventConsumer(consumerCfg, dbPool)
	ctx, cancel := context.WithCancel(context.Background())
	go consumer.Start(ctx)

	// Маршруты
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", homeHandler)
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /shorten", urlHandlers.Shorten)
	mux.HandleFunc("GET /r/{code}", urlHandlers.Redirect)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		log.Println("Server started on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Ждём сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Останавливаем consumer
	cancel()
	// Даём consumer время завершить обработку
	time.Sleep(2 * time.Second)

	// Завершаем HTTP сервер
	ctxShutdown, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("URL Shortener Service"))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
