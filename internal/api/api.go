package api

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/vscodethemes/backend/internal/api/handlers"
	"github.com/vscodethemes/backend/internal/api/middleware"
)

func Config() huma.Config {
	config := huma.DefaultConfig("VS Code Themes API", "1.0.0")

	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{}
	config.Components.SecuritySchemes[middleware.BearerAuthSecurityKey] = &huma.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}

	return config
}

func RegisterRoutes(api huma.API, publicKeyPath string, issuer string, h handlers.Handler) {
	api.UseMiddleware(middleware.Auth(api, publicKeyPath, issuer))

	huma.Register(api, handlers.ScanExtensionsOperation, h.ScanExtensions)
	huma.Register(api, handlers.GetExtensionOperation, h.GetExtension)
	huma.Register(api, handlers.SyncExtensionOperation, h.SyncExtension)
	huma.Register(api, handlers.GetJobOperation, h.GetJob)
}
