package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config хранит настройки подключения к базе данных
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// NewPool создаёт пул соединений с PostgreSQL и проверяет подключение
func NewPool(cfg Config) (*pgxpool.Pool, error) {
	// Формируем строку подключения (DSN)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	// Парсим конфигурацию пула из DSN
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	// Настраиваем параметры пула (для продакшна)
	config.MaxConns = 10                      // максимум 10 соединений
	config.MinConns = 2                       // минимум 2 соединения в запасе
	config.MaxConnLifetime = time.Hour        // максимум время жизни соединения
	config.MaxConnIdleTime = 30 * time.Minute // максимум простоя

	// Создаём пул соединений
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Пробуем выполнить ping, чтобы убедиться, что БД доступна
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL")
	return pool, nil
}
