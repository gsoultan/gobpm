package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-kit/kit/endpoint"
	"github.com/gsoultan/gobpm/internal/pkg/auth"
	"github.com/gsoultan/gobpm/server/interceptors/contracts"
)

// endpointAuthInterceptor verifies the JWT token from context (extracted in transport).
type endpointAuthInterceptor struct{}

func NewEndpointAuthInterceptor() contracts.EndpointInterceptor {
	return &endpointAuthInterceptor{}
}

func (i *endpointAuthInterceptor) Intercept(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		user := ctx.Value(auth.UserContextKey)
		if user == nil {
			return nil, auth.ErrUnauthorized
		}
		return next(ctx, request)
	}
}

// httpAuthInterceptor extracts the user from the Authorization header and puts it in the context.
type httpAuthInterceptor struct {
	strategy SecurityStrategy
}

func NewHTTPAuthInterceptor(strategy SecurityStrategy) contracts.TransportInterceptor {
	return &httpAuthInterceptor{strategy: strategy}
}

func (i *httpAuthInterceptor) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			next.ServeHTTP(w, r)
			return
		}

		u, err := i.strategy.Authenticate(r.Context(), parts[1])
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), auth.UserContextKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// mandatoryHTTPAuthInterceptor is like httpAuthInterceptor but returns 401 on failure.
type mandatoryHTTPAuthInterceptor struct {
	strategy    SecurityStrategy
	publicPaths []string
}

func NewMandatoryHTTPAuthInterceptor(strategy SecurityStrategy, publicPaths []string) contracts.TransportInterceptor {
	return &mandatoryHTTPAuthInterceptor{
		strategy:    strategy,
		publicPaths: publicPaths,
	}
}

func (i *mandatoryHTTPAuthInterceptor) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only protect /api/ endpoints.
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		// Exclude public API endpoints
		for _, path := range i.publicPaths {
			if r.URL.Path == path {
				next.ServeHTTP(w, r)
				return
			}
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid auth header", http.StatusUnauthorized)
			return
		}

		u, err := i.strategy.Authenticate(r.Context(), parts[1])
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), auth.UserContextKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
