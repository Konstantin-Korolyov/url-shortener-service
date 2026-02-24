package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/Konstantin-Korolyov/url-shortener-go/internal/cache"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/kafka"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/middleware"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/models"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/repository"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/utils"
)

type URLHandlers struct {
	repo     *repository.URLRepository
	cache    *cache.RedisClient
	producer *kafka.Producer // новый продюсер
}

func NewURLHandlers(repo *repository.URLRepository, cache *cache.RedisClient, producer *kafka.Producer) *URLHandlers {
	return &URLHandlers{repo: repo, cache: cache, producer: producer}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func (h *URLHandlers) Shorten(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод
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

	// --- Извлекаем user_id из контекста (если есть) ---
	var userID *int64
	if uid, ok := r.Context().Value(middleware.UserIDKey).(int64); ok {
		userID = &uid
	}

	const maxAttempts = 5
	var url *models.URL
	var err error

	for i := 0; i < maxAttempts; i++ {
		shortCode := utils.RandomString(6)
		url = &models.URL{
			OriginalURL: req.URL,
			ShortCode:   shortCode,
			UserID:      userID, // теперь заполняем, если есть
			Clicks:      0,
			CreatedAt:   time.Now(),
			ExpiresAt:   nil,
			IsActive:    true,
		}
		err = h.repo.Create(r.Context(), url)
		if err == nil {
			break
		}
		// Проверяем на уникальность
		if repository.IsUniqueViolation(err) {
			slog.Warn("Short code collision, retrying", "code", shortCode, "attempt", i+1)
			continue
		}
		break
	}

	if err != nil {
		slog.Error("Failed to create URL after attempts", "err", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Сохраняем в кеш
	if err := h.cache.SetURL(r.Context(), url.ShortCode, url, time.Hour); err != nil {
		slog.Error("Failed to cache URL", "err", err)
	}

	// Ответ
	resp := shortenResponse{
		ShortURL: "http://localhost:8080/r/" + url.ShortCode,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *URLHandlers) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	// 1. Пытаемся получить из кеша
	url, err := h.cache.GetURL(r.Context(), code)
	if err != nil {
		slog.Error("Redis error", "err", err)
	}
	if url != nil {
		if !url.IsActive || (url.ExpiresAt != nil && url.ExpiresAt.Before(time.Now())) {
			http.Error(w, "URL expired or inactive", http.StatusGone)
			return
		}
		// Увеличиваем счётчик
		if err := h.repo.IncrementClicks(r.Context(), url.ID); err != nil {
			slog.Error("Failed to increment clicks", "err", err)
		}

		// Отправляем событие в Kafka (асинхронно)
		if h.producer != nil {
			event := kafka.ClickEvent{
				URLID:     url.ID,
				IP:        func() string { h, _, _ := net.SplitHostPort(r.RemoteAddr); return h }(),
				UserAgent: r.UserAgent(),
				Referer:   r.Referer(),
				Timestamp: time.Now(),
			}
			go func() {
				if err := h.producer.PublishClick(context.Background(), event); err != nil {
					slog.Error("Failed to publish click event", "err", err)
				}
			}()
		}

		http.Redirect(w, r, url.OriginalURL, http.StatusFound)
		return
	}

	// 2. Если в кеше нет, ищем в БД
	url, err = h.repo.FindByShortCode(r.Context(), code)
	if err != nil {
		slog.Error("DB error", "err", err)
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
		slog.Error("Failed to cache URL", "err", err)
	}

	// Увеличиваем счётчик
	if err := h.repo.IncrementClicks(r.Context(), url.ID); err != nil {
		slog.Error("Failed to increment clicks", "err", err)
	}

	// Отправляем событие в Kafka (асинхронно)
	if h.producer != nil {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		event := kafka.ClickEvent{
			URLID:     url.ID,
			IP:        ip,
			UserAgent: r.UserAgent(),
			Referer:   r.Referer(),
			Timestamp: time.Now(),
		}
		go func() {
			if err := h.producer.PublishClick(context.Background(), event); err != nil {
				slog.Error("Failed to publish click event", "err", err)
			}
		}()
	}

	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}
