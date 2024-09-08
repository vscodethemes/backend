package api

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/vscodethemes/backend/internal/api/handlers"
)

func RegisterRoutes(api huma.API, h handlers.Handler) {
	huma.Register(api, handlers.GetExtensionOperation, h.GetExtension)
	huma.Register(api, handlers.SyncExtensionOperation, h.SyncExtension)
	huma.Register(api, handlers.GetJobOperation, h.GetJob)
}
