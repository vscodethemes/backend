package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	middleware "github.com/oapi-codegen/echo-middleware"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/vscodethemes/backend/internal/api"
)

func main() {
	port := flag.String("port", "8080", "Port for test HTTP server")
	dbUrl := flag.String("db-url", "", "Database URL")
	flag.Parse()

	if *dbUrl == "" {
		fmt.Fprintf(os.Stderr, "Database URL is required\n")
		os.Exit(1)
	}

	fmt.Println("dbUrl", *dbUrl)

	dbPool, err := pgxpool.New(context.Background(), *dbUrl)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create db pool: %w", err))
	}
	defer dbPool.Close()

	// Insert-only river client.
	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{})
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create river client: %w", err))
	}

	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	server := api.NewServer(dbPool, riverClient)

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(middleware.OapiRequestValidator(swagger))

	api.RegisterHandlers(e, server)

	// TODO: Graceful shutdown.
	// https://echo.labstack.com/docs/cookbook/graceful-shutdown

	// And we serve HTTP until the world ends.
	e.Logger.Fatal(e.Start(net.JoinHostPort("0.0.0.0", *port)))
}
