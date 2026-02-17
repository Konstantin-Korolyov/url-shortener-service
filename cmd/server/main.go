package main

import (
	"log"
	"net/http"

	"github.com/Konstantin-Korolyov/url-shortener-go/internal/cache"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/database"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/handlers"
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

	// Репозиторий
	urlRepo := repository.NewURLRepository(dbPool)

	// Обработчики (создадим структуру в следующем шаге)
	urlHandlers := handlers.NewURLHandlers(urlRepo, redisClient)

	// Маршруты
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", homeHandler)
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /shorten", urlHandlers.Shorten)
	mux.HandleFunc("GET /r/{code}", urlHandlers.Redirect)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("URL Shortener Service"))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
