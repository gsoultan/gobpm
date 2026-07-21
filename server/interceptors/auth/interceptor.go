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

		token, ok := bearerTokenFromHeader(authHeader)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		u, err := i.strategy.Authenticate(r.Context(), token)
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
	strategy      SecurityStrategy
	publicPathSet map[string]struct{}
}

func NewMandatoryHTTPAuthInterceptor(strategy SecurityStrategy, publicPaths []string) contracts.TransportInterceptor {
	publicPathSet := make(map[string]struct{}, len(publicPaths))
	for _, publicPath := range publicPaths {
		publicPathSet[publicPath] = struct{}{}
	}

	return &mandatoryHTTPAuthInterceptor{
		strategy:      strategy,
		publicPathSet: publicPathSet,
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
		if i.isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, ok := bearerTokenFromHeader(authHeader)
		if !ok {
			http.Error(w, "Invalid auth header", http.StatusUnauthorized)
			return
		}

		u, err := i.strategy.Authenticate(r.Context(), token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), auth.UserContextKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (i *mandatoryHTTPAuthInterceptor) isPublicPath(path string) bool {
	_, ok := i.publicPathSet[path]
	return ok
}

func bearerTokenFromHeader(authHeader string) (string, bool) {
	prefix, token, found := strings.Cut(authHeader, " ")
	if !found || prefix != "Bearer" || token == "" || strings.ContainsAny(token, " \t") {
		return "", false
	}

	return token, true
}
