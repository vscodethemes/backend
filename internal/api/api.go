package api

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/vscodethemes/backend/internal/api/handlers"
)

func RegisterRoutes(api huma.API, h handlers.Handler) {
	huma.Register(api, handlers.SyncExtensionBySlugOperation, h.SyncExtensionBySlug)
	huma.Register(api, handlers.GetJobByIDOperation, h.GetJobByID)
}
