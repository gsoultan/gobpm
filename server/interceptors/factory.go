package interceptors

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/gsoultan/gobpm/internal/pkg/auth"
	"github.com/gsoultan/gobpm/server/domains/services"
	authinterceptor "github.com/gsoultan/gobpm/server/interceptors/auth"
	"github.com/gsoultan/gobpm/server/interceptors/contracts"
	"github.com/gsoultan/gobpm/server/interceptors/logging"
)

// InterceptorFactory creates various interceptors.
type InterceptorFactory struct {
	svc services.ServiceFacade
}

func NewInterceptorFactory(svc services.ServiceFacade) *InterceptorFactory {
	return &InterceptorFactory{svc: svc}
}

func (f *InterceptorFactory) NewLogging(method string) contracts.EndpointInterceptor {
	return logging.NewLoggingInterceptor(method)
}

func (f *InterceptorFactory) NewEndpointAuth() contracts.EndpointInterceptor {
	return authinterceptor.NewEndpointAuthInterceptor()
}

func (f *InterceptorFactory) NewHTTPAuth(strategy authinterceptor.SecurityStrategy) contracts.TransportInterceptor {
	return authinterceptor.NewHTTPAuthInterceptor(strategy)
}

func (f *InterceptorFactory) NewMandatoryHTTPAuth(strategy authinterceptor.SecurityStrategy, publicPaths []string) contracts.TransportInterceptor {
	return authinterceptor.NewMandatoryHTTPAuthInterceptor(strategy, publicPaths)
}

func (f *InterceptorFactory) NewJWTStrategy() authinterceptor.SecurityStrategy {
	return authinterceptor.NewJWTStrategy(f.svc.ValidateToken)
}

func (f *InterceptorFactory) NewOIDCStrategy(validator *auth.TokenValidator) authinterceptor.SecurityStrategy {
	return authinterceptor.NewOIDCStrategy(func(ctx context.Context, token string) (any, error) {
		return validator.ValidateToken(ctx, token)
	})
}

// ProtectedChain returns a function that applies both logging and auth to an endpoint.
func (f *InterceptorFactory) ProtectedChain(method string) func(endpoint.Endpoint) endpoint.Endpoint {
	logging := f.NewLogging(method)
	auth := f.NewEndpointAuth()
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return auth.Intercept(logging.Intercept(e))
	}
}

// PublicChain returns a function that applies only logging to an endpoint.
func (f *InterceptorFactory) PublicChain(method string) func(endpoint.Endpoint) endpoint.Endpoint {
	logging := f.NewLogging(method)
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return logging.Intercept(e)
	}
}
