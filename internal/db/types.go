package db

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func Timestamp(t *time.Time) pgtype.Timestamp {
	timestamp := pgtype.Timestamp{}

	if t != nil {
		timestamp.Time = *t
		timestamp.Valid = true
	}

	return timestamp
}

func Text(s *string) pgtype.Text {
	text := pgtype.Text{}

	if s != nil {
		text.String = *s
		text.Valid = true
	}

	return text
}

func Numeric(n *float64) (pgtype.Numeric, error) {
	numeric := pgtype.Numeric{}

	if n != nil {
		strFloat := strconv.FormatFloat(*n, 'f', -1, 64)
		err := numeric.Scan(strFloat)
		if err != nil {
			return numeric, fmt.Errorf("failed to scan numeric: %w", err)
		}
	}

	return numeric, nil
}
