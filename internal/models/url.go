package models

import (
	"time"
)

// URL представляет запись в таблице urls
type URL struct {
	ID          int64      `json:"id"`                   // первичный ключ
	OriginalURL string     `json:"original_url"`         // оригинальная длинная ссылка
	ShortCode   string     `json:"short_code"`           // короткий код (уникальный)
	UserID      *int64     `json:"user_id,omitempty"`    // ID пользователя (может быть NULL)
	Clicks      int64      `json:"clicks"`               // счётчик переходов
	CreatedAt   time.Time  `json:"created_at"`           // время создания
	ExpiresAt   *time.Time `json:"expires_at,omitempty"` // время истечения (NULL если не указано)
	IsActive    bool       `json:"is_active"`            // активна ли ссылка
}
