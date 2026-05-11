package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"honeygarden/internal/domain"
)

func isUniqueViolation(err error, constraint string) bool {
	if err == nil {
		return false
	}
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) {
		return false
	}
	if pgError.Code != "23505" {
		return false
	}
	return constraint == "" || pgError.ConstraintName == constraint
}

func pgErr(err error) error {
	if err == nil {
		return nil
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.Code {
		case "23505":
			return domain.ErrAlreadyExists
		case "23503":
			return domain.ErrNotFound
		}
	}
	return err
}
