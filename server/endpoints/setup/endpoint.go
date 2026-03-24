package setup

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/gsoultan/gobpm/server/domains/services/contracts"
)

type Endpoints struct {
	GetSetupStatusEndpoint endpoint.Endpoint
	SetupEndpoint          endpoint.Endpoint
	TestConnectionEndpoint endpoint.Endpoint
}

func MakeEndpoints(s contracts.SetupService) Endpoints {
	return Endpoints{
		GetSetupStatusEndpoint: MakeGetSetupStatusEndpoint(s),
		SetupEndpoint:          MakeSetupEndpoint(s),
		TestConnectionEndpoint: MakeTestConnectionEndpoint(s),
	}
}

func MakeGetSetupStatusEndpoint(s contracts.SetupService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		status, err := s.GetSetupStatus(ctx)
		return GetSetupStatusResponse{Status: status, Err: err}, nil
	}
}

func MakeSetupEndpoint(s contracts.SetupService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(SetupRequest)
		err := s.Setup(ctx, contracts.SetupRequest{
			AdminUsername:    req.AdminUsername,
			AdminPassword:    req.AdminPassword,
			AdminFullName:    req.AdminFullName,
			AdminPublicName:  req.AdminPublicName,
			AdminEmail:       req.AdminEmail,
			OrganizationName: req.OrganizationName,
			ProjectName:      req.ProjectName,
			DatabaseDriver:   req.DatabaseDriver,
			DBHost:           req.DBHost,
			DBPort:           req.DBPort,
			DBUsername:       req.DBUsername,
			DBPassword:       req.DBPassword,
			DBName:           req.DBName,
			DBSSLEnabled:     req.DBSSLEnabled,
			EncryptionKey:    req.EncryptionKey,
			JWTSecret:        req.JWTSecret,
		})
		return SetupResponse{Err: err}, nil
	}
}

func MakeTestConnectionEndpoint(s contracts.SetupService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(TestConnectionRequest)
		result := s.TestConnection(ctx, contracts.TestConnectionRequest{
			DatabaseDriver: req.DatabaseDriver,
			DBHost:         req.DBHost,
			DBPort:         req.DBPort,
			DBUsername:     req.DBUsername,
			DBPassword:     req.DBPassword,
			DBName:         req.DBName,
			DBSSLEnabled:   req.DBSSLEnabled,
		})
		return TestConnectionResponse{Success: result.Success, Message: result.Message}, nil
	}
}
