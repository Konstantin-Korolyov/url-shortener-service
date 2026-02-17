package main

import (
	"log"
	"net/http"

	"github.com/Konstantin-Korolyov/url-shortener-go/internal/database"
)

func main() {
	// Настройки подключения к БД (пока захардкожены)
	dbCfg := database.Config{
		Host:     "localhost", // потому что сервер запущен на хосте, а контейнер слушает localhost:5432
		Port:     "5432",
		User:     "admin",
		Password: "securepassword123",
		DBName:   "shortener_db",
	}

	// Подключаемся к БД
	dbPool, err := database.NewPool(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Закрываем пул при завершении программы
	defer dbPool.Close()

	// Простейшие обработчики (пока не используют БД)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/r/", redirectHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("URL Shortener Service"))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Shorten endpoint (not implemented)"))
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Redirect endpoint (not implemented)"))
}
