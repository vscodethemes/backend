package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/vscodethemes/backend/internal/api"
	"github.com/vscodethemes/backend/internal/api/handlers"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
	"github.com/danielgtaylor/huma/v2/humacli"
)

// Options for the CLI.
type Options struct {
	Host          string `help:"Host to listen on" default:"0.0.0.0"`
	Port          int    `help:"Port to listen on" default:"8080"`
	DatabaseURL   string `help:"Database URL" required:"true"`
	PublicKeyPath string `help:"Path to the public key file" default:"key.rsa.pub"`
	Issuer        string `help:"JWT issuer" default:"localhost:8080"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create a CLI app which takes a port option.
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Create a new database pool.
		dbPool, err := pgxpool.New(context.Background(), options.DatabaseURL)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to create database pool: %s", err))
			os.Exit(1)
		}

		// Create insert-only river client.
		riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
			Logger: logger,
		})
		if err != nil {
			logger.Error(fmt.Sprintf("failed to create river client: %s", err))
			os.Exit(1)
		}

		// Create a new API server.
		server := api.NewServer(logger, options.PublicKeyPath, options.Issuer, handlers.Handler{
			DBPool:      dbPool,
			RiverClient: riverClient,
			Logger:      logger,
		})

		// Tell the CLI how to start your server.
		hooks.OnStart(func() {
			port := fmt.Sprintf("%d", options.Port)
			if err := server.Start(net.JoinHostPort(options.Host, port)); err != nil {
				logger.Error(fmt.Sprintf("failed to start server: %s", err))
				os.Exit(1)
			}
		})

		hooks.OnStop(func() {
			// This will block until all queries are finished. There does not seem to be a way to
			// set a timeout on this. This comment from the author of pgx suggests to just terminate
			// the application and let postgres handle the connection cleanup:
			//  https://github.com/jackc/pgx/issues/802#issuecomment-668713840
			// defer dbPool.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			logger.Info("Shutting down server")
			if err := server.Shutdown(ctx); err != nil {
				logger.Error(fmt.Sprintf("failed to shutdown server: %s", err))
				os.Exit(1)
			}
		})
	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
