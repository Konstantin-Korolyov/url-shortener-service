package repository_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Konstantin-Korolyov/url-shortener-go/internal/database"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/models"
	"github.com/Konstantin-Korolyov/url-shortener-go/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	// Поднимаем PostgreSQL в Docker
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("postgres", "15-alpine", []string{
		"POSTGRES_USER=test",
		"POSTGRES_PASSWORD=test",
		"POSTGRES_DB=testdb",
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// Ждём, пока БД поднимется
	if err := pool.Retry(func() error {
		var err error
		testDB, err = database.NewPool(database.Config{
			Host:     "localhost",
			Port:     resource.GetPort("5432/tcp"),
			User:     "test",
			Password: "test",
			DBName:   "testdb",
		})
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Накатываем схему из init.sql
	initSQL, err := os.ReadFile("../../docker/postgres/init.sql")
	if err != nil {
		log.Fatal("Failed to read init.sql:", err)
	}
	if _, err := testDB.Exec(context.Background(), string(initSQL)); err != nil {
		log.Fatal("Failed to execute init.sql:", err)
	}

	code := m.Run()

	// Очистка
	_ = pool.Purge(resource)
	os.Exit(code)
}

func TestCreateAndFind(t *testing.T) {
	repo := repository.NewURLRepository(testDB)
	ctx := context.Background()

	url := &models.URL{
		OriginalURL: "https://example.com",
		ShortCode:   "test123",
		Clicks:      0,
		CreatedAt:   time.Now(),
		IsActive:    true,
	}
	err := repo.Create(ctx, url)
	assert.NoError(t, err)
	assert.NotZero(t, url.ID)

	found, err := repo.FindByShortCode(ctx, "test123")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, url.OriginalURL, found.OriginalURL)

	// Проверка, что не найден
	notFound, err := repo.FindByShortCode(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, notFound)
}
