package repository

import (
	"context"
	"errors"

	"github.com/Konstantin-Korolyov/url-shortener-go/internal/models" // замени на свой путь
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// URLRepository предоставляет методы для работы с таблицей urls
type URLRepository struct {
	db *pgxpool.Pool
}

// NewURLRepository создаёт новый экземпляр репозитория
func NewURLRepository(db *pgxpool.Pool) *URLRepository {
	return &URLRepository{db: db}
}

// Create добавляет новую запись в таблицу urls
// Возвращает заполненную структуру с ID и временем создания (если нужно)
func (r *URLRepository) Create(ctx context.Context, url *models.URL) error {
	query := `
		INSERT INTO urls (original_url, short_code, user_id, clicks, created_at, expires_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query,
		url.OriginalURL,
		url.ShortCode,
		url.UserID,
		url.Clicks,
		url.CreatedAt,
		url.ExpiresAt,
		url.IsActive,
	).Scan(&url.ID, &url.CreatedAt) // получаем сгенерированные значения
	if err != nil {
		return err
	}
	return nil
}

// FindByShortCode ищет URL по короткому коду
// Если запись не найдена, возвращает (nil, nil)
func (r *URLRepository) FindByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	query := `
		SELECT id, original_url, short_code, user_id, clicks, created_at, expires_at, is_active
		FROM urls
		WHERE short_code = $1
	`
	row := r.db.QueryRow(ctx, query, shortCode)

	var url models.URL
	err := row.Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.UserID,
		&url.Clicks,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &url, nil
}

// IncrementClicks увеличивает счётчик кликов для указанного ID
func (r *URLRepository) IncrementClicks(ctx context.Context, id int64) error {
	query := `UPDATE urls SET clicks = clicks + 1 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
