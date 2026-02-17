package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Konstantin-Korolyov/url-shortener-go/internal/cache"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/models"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/repository"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/utils"
)

type URLHandlers struct {
	repo  *repository.URLRepository
	cache *cache.RedisClient
}

func NewURLHandlers(repo *repository.URLRepository, cache *cache.RedisClient) *URLHandlers {
	return &URLHandlers{repo: repo, cache: cache}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func (h *URLHandlers) Shorten(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод (хотя mux уже должен был это сделать, но для надёжности)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Декодируем JSON
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Генерируем короткий код (6 символов)
	shortCode := utils.RandomString(6)

	// Создаём модель URL
	url := &models.URL{
		OriginalURL: req.URL,
		ShortCode:   shortCode,
		UserID:      nil, // пока без пользователя
		Clicks:      0,
		CreatedAt:   time.Now(),
		ExpiresAt:   nil, // пока без срока
		IsActive:    true,
	}

	// Сохраняем в БД
	if err := h.repo.Create(r.Context(), url); err != nil {
		log.Printf("Failed to create URL: %v", err)
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	// Сохраняем в кеш (на 1 час) для быстрого доступа
	if err := h.cache.SetURL(r.Context(), shortCode, url, time.Hour); err != nil {
		log.Printf("Failed to cache URL: %v", err) // только логируем, не прерываем ответ
	}

	// Формируем ответ
	resp := shortenResponse{
		ShortURL: "http://localhost:8080/r/" + shortCode,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *URLHandlers) Redirect(w http.ResponseWriter, r *http.Request) {
	// Извлекаем код из пути
	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	// 1. Пытаемся получить из кеша
	url, err := h.cache.GetURL(r.Context(), code)
	if err != nil {
		log.Printf("Redis error: %v", err)
		// Если ошибка Redis, продолжаем без кеша
	}
	if url != nil {
		// Проверяем, активна ли ссылка и не истекла ли
		if !url.IsActive || (url.ExpiresAt != nil && url.ExpiresAt.Before(time.Now())) {
			http.Error(w, "URL expired or inactive", http.StatusGone)
			return
		}
		// Асинхронно увеличиваем счётчик? Пока сделаем синхронно
		if err := h.repo.IncrementClicks(r.Context(), url.ID); err != nil {
			log.Printf("Failed to increment clicks: %v", err)
		}
		http.Redirect(w, r, url.OriginalURL, http.StatusFound)
		return
	}

	// 2. Если в кеше нет, ищем в БД
	url, err = h.repo.FindByShortCode(r.Context(), code)
	if err != nil {
		log.Printf("DB error: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if url == nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	if !url.IsActive {
		http.Error(w, "URL is inactive", http.StatusGone)
		return
	}
	if url.ExpiresAt != nil && url.ExpiresAt.Before(time.Now()) {
		http.Error(w, "URL expired", http.StatusGone)
		return
	}

	// Сохраняем в кеш для будущих запросов
	if err := h.cache.SetURL(r.Context(), code, url, time.Hour); err != nil {
		log.Printf("Failed to cache URL: %v", err)
	}

	// Увеличиваем счётчик
	if err := h.repo.IncrementClicks(r.Context(), url.ID); err != nil {
		log.Printf("Failed to increment clicks: %v", err)
	}

	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}
