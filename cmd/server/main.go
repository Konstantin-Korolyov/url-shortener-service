package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Konstantin-Korolyov/url-shortener-go/internal/cache"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/config"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/database"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/handlers"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/kafka"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/middleware"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/repository"
)

func main() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	// Если нужен текст, можно заменить на:
	// handler := slog.NewTextHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(handler))

	cfg := config.Load()

	// PostgreSQL
	dbCfg := database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	}
	dbPool, err := database.NewPool(dbCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	// Redis
	redisAddr := cfg.RedisHost + ":" + cfg.RedisPort
	redisClient, err := cache.NewRedisClient(redisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Kafka producer
	producer := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer producer.Close()

	// Репозиторий
	urlRepo := repository.NewURLRepository(dbPool)

	// Обработчики
	urlHandlers := handlers.NewURLHandlers(urlRepo, redisClient, producer)

	// Kafka consumer
	consumerCfg := kafka.ConsumerConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTopic,
		GroupID: "click-consumers",
	}
	consumer := kafka.NewClickEventConsumer(consumerCfg, dbPool)
	ctx, cancel := context.WithCancel(context.Background())
	go consumer.Start(ctx)

	// HTTP сервер
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", homeHandler)
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /shorten", urlHandlers.Shorten)
	mux.HandleFunc("GET /r/{code}", urlHandlers.Redirect)

	// Оборачиваем весь маршрутизатор в rate limiter
	rateLimitedHandler := middleware.RateLimit(mux)
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: rateLimitedHandler,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server started on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	cancel()
	time.Sleep(2 * time.Second)

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
