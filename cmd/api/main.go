package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

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

		// TODO: Graceful shutdown.
		// https://echo.labstack.com/docs/cookbook/graceful-shutdown
		// https://huma.rocks/how-to/graceful-shutdown

		// Tell the CLI how to start your server.
		hooks.OnStart(func() {
			port := fmt.Sprintf("%d", options.Port)
			if err := server.Start(net.JoinHostPort(options.Host, port)); err != nil {
				logger.Error(fmt.Sprintf("failed to start server: %s", err))
				os.Exit(1)
			}
		})

		hooks.OnStop(func() {
			dbPool.Close()
		})
	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
