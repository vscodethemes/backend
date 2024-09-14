package middleware

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var BearerAuthSecurityKey = "bearerAuth"

func BearerAuthSecurity(anyOfNeededScopes ...string) map[string][]string {
	security := map[string][]string{}

	if len(anyOfNeededScopes) > 0 {
		security[BearerAuthSecurityKey] = anyOfNeededScopes
	}

	return security
}

// Auth creates a middleware that will authorize requests based on the required scopes for the operation.
func Auth(api huma.API, publicKeyPath string, issuer string) func(ctx huma.Context, next func(huma.Context)) {
	// Read public key from file.
	keyPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read public key from file: %w", err))
	}

	// Parse public key.
	publicKey, err := jwk.ParseKey(keyPEM, jwk.WithPEM(true))
	if err != nil {
		panic(fmt.Errorf("failed to parse public key: %w", err))
	}

	return func(ctx huma.Context, next func(huma.Context)) {
		// Check if the operation requires authorization and extract the required scopes.
		var anyOfNeededScopes []string
		isAuthorizationRequired := false
		for _, opScheme := range ctx.Operation().Security {
			var ok bool
			if anyOfNeededScopes, ok = opScheme[BearerAuthSecurityKey]; ok {
				isAuthorizationRequired = true
				break
			}
		}

		if !isAuthorizationRequired {
			next(ctx)
			return
		}

		// Extract the JWT token from the Authorization header.
		token := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
		if len(token) == 0 {
			mustWriteError(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Parse and validate the JWT.
		parsed, err := jwt.ParseString(token,
			jwt.WithKey(jwa.RS256, publicKey),
			jwt.WithIssuer(issuer),
		)

		if err != nil {
			mustWriteError(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Ensure the claims required for this operation are present.
		scopes, _ := parsed.Get("scopes")
		if scopes, ok := scopes.([]interface{}); ok {
			for _, scope := range scopes {
				if scope, ok := scope.(string); ok && slices.Contains(anyOfNeededScopes, scope) {
					next(ctx)
					return
				}
			}
		}

		mustWriteError(api, ctx, http.StatusForbidden, "Forbidden")
	}
}

func mustWriteError(api huma.API, ctx huma.Context, status int, message string) {
	err := huma.WriteErr(api, ctx, status, message)
	if err != nil {
		panic(fmt.Errorf("failed to write error: %w", err))
	}
}
