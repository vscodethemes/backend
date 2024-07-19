package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/vscodethemes/backend/internal/api"
	"github.com/vscodethemes/backend/internal/api/handlers"

	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
	"github.com/danielgtaylor/huma/v2/humacli"
)

// Options for the CLI.
type Options struct {
	Host        string `help:"Host to listen on" default:"0.0.0.0"`
	Port        int    `help:"Port to listen on" default:"8080"`
	DatabaseURL string `help:"Database URL" required:"true"`
}

func main() {
	// Create a CLI app which takes a port option.
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Create a new database pool.
		dbPool, err := pgxpool.New(context.Background(), options.DatabaseURL)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to create db pool: %w", err))
		}

		// Create insert-only river client.
		riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{})
		if err != nil {
			log.Fatal(fmt.Errorf("failed to create river client: %w", err))
		}

		// Create a new router & API
		e := echo.New()
		e.Use(echomiddleware.Logger())
		humaApi := humaecho.New(e, huma.DefaultConfig("VS Code Themes API", "1.0.0"))

		api.RegisterRoutes(humaApi, handlers.Handler{
			DBPool:      dbPool,
			RiverClient: riverClient,
		})

		// TODO: Graceful shutdown.
		// https://echo.labstack.com/docs/cookbook/graceful-shutdown
		// https://huma.rocks/how-to/graceful-shutdown

		// Tell the CLI how to start your server.
		hooks.OnStart(func() {
			port := fmt.Sprintf("%d", options.Port)
			e.Logger.Fatal(e.Start(net.JoinHostPort(options.Host, port)))
		})

		hooks.OnStop(func() {
			dbPool.Close()
		})
	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
