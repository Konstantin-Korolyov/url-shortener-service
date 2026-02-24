package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolation = "23505"

// IsUniqueViolation возвращает true, если ошибка связана с нарушением уникальности.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == uniqueViolation
	}
	return false
}
