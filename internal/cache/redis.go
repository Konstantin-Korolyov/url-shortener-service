package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Konstantin-Korolyov/url-shortener-go/internal/models" // замени на свой путь
	"github.com/redis/go-redis/v9"
)

// RedisClient обёртка над redis.Client, чтобы добавить свои методы
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient создаёт новый клиент Redis и проверяет подключение
func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{client: client}, nil
}

// SetURL сохраняет URL в кеш с указанным временем жизни
func (r *RedisClient) SetURL(ctx context.Context, shortCode string, url *models.URL, expiration time.Duration) error {
	// Сериализуем структуру в JSON
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	// Сохраняем по ключу "url:" + shortCode
	return r.client.Set(ctx, "url:"+shortCode, data, expiration).Err()
}

// GetURL получает URL из кеша по короткому коду
// Возвращает (nil, nil) если ключ не найден
func (r *RedisClient) GetURL(ctx context.Context, shortCode string) (*models.URL, error) {
	data, err := r.client.Get(ctx, "url:"+shortCode).Bytes()
	if err == redis.Nil {
		// Ключа нет в кеше – это не ошибка, а промах
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Десериализуем JSON в структуру
	var url models.URL
	if err := json.Unmarshal(data, &url); err != nil {
		return nil, err
	}
	return &url, nil
}

// DeleteURL удаляет URL из кеша (на всякий случай)
func (r *RedisClient) DeleteURL(ctx context.Context, shortCode string) error {
	return r.client.Del(ctx, "url:"+shortCode).Err()
}
