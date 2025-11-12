package dbsamples

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func IsNotFound(err error) bool {
	if errors.Is(err, pgx.ErrNoRows) {
		return true
	}
	return IsCode(err, "20000")
}

func IsDuplicated(err error) bool {
	return IsCode(err, "23505")
}

func IsCode(err error, code string) bool {
	pgErr := AsPgError(err)
	return pgErr != nil && pgErr.Code == code
}

func AsPgError(err error) *pgconn.PgError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr
	}
	return nil
}
