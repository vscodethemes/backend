package api

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/vscodethemes/backend/internal/api/handlers"
	"github.com/vscodethemes/backend/internal/api/middleware"
)

func NewServer(logger *slog.Logger, publicKeyPath string, issuer string, h handlers.Handler) *echo.Echo {
	// Echo-specific setup.
	e := echo.New()
	e.Use(echomiddleware.RequestID())
	e.Use(middleware.Logger(logger))

	// Huma-specific setup.
	config := huma.DefaultConfig("VS Code Themes API", "1.0.0")
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{}
	config.Components.SecuritySchemes[middleware.BearerAuthSecurityKey] = &huma.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}

	api := humaecho.New(e, config)
	api.UseMiddleware(middleware.Auth(api, publicKeyPath, issuer))

	// Register routes.
	huma.Register(api, handlers.SearchExtensionsOperation, h.SearchExtensions)
	huma.Register(api, handlers.ScanExtensionsOperation, h.ScanExtensions)
	huma.Register(api, handlers.SyncExtensionOperation, h.SyncExtension)
	huma.Register(api, handlers.GetJobOperation, h.GetJob)
	huma.Register(api, handlers.PauseJobsOperation, h.PauseJobs)
	huma.Register(api, handlers.ResumeJobsOperation, h.ResumeJobs)
	huma.Register(api, handlers.GetColorsOperation, h.GetColors)

	return e
}
