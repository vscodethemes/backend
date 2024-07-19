package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Tx(ctx context.Context, db *pgxpool.Pool, txFunc func(pgx.Tx) error) (err error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				fmt.Printf("failed to rollback transaction: %v\n", err)
			}
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			err = tx.Rollback(ctx) // err is non-nil; don't change it
			if err != nil {
				fmt.Printf("failed to rollback transaction: %v\n", err)
			}
		} else {
			err = tx.Commit(ctx) // err is nil; if Commit returns error update err
			if err != nil {
				fmt.Printf("failed to commit transaction: %v\n", err)
			}
		}
	}()
	err = txFunc(tx)
	return err
}
