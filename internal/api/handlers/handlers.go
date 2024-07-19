package handlers

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

type Handler struct {
	DBPool      *pgxpool.Pool
	RiverClient *river.Client[pgx.Tx]
}
