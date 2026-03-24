package contracts

import (
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

// EndpointInterceptor defines the contract for endpoint-level cross-cutting concerns.
type EndpointInterceptor interface {
	Intercept(next endpoint.Endpoint) endpoint.Endpoint
}

// TransportInterceptor defines the contract for transport-level (HTTP/gRPC) middlewares.
type TransportInterceptor interface {
	Wrap(next http.Handler) http.Handler
}
