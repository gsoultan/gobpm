package connector

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/services"
)

type Endpoints struct {
	ListConnectors          endpoint.Endpoint
	CreateConnector         endpoint.Endpoint
	UpdateConnector         endpoint.Endpoint
	DeleteConnector         endpoint.Endpoint
	ListConnectorInstances  endpoint.Endpoint
	CreateConnectorInstance endpoint.Endpoint
	UpdateConnectorInstance endpoint.Endpoint
	DeleteConnectorInstance endpoint.Endpoint
	ExecuteConnector        endpoint.Endpoint
}

func MakeEndpoints(s services.ServiceFacade) Endpoints {
	return Endpoints{
		ListConnectors:          MakeListConnectorsEndpoint(s),
		CreateConnector:         MakeCreateConnectorEndpoint(s),
		UpdateConnector:         MakeUpdateConnectorEndpoint(s),
		DeleteConnector:         MakeDeleteConnectorEndpoint(s),
		ListConnectorInstances:  MakeListConnectorInstancesEndpoint(s),
		CreateConnectorInstance: MakeCreateConnectorInstanceEndpoint(s),
		UpdateConnectorInstance: MakeUpdateConnectorInstanceEndpoint(s),
		DeleteConnectorInstance: MakeDeleteConnectorInstanceEndpoint(s),
		ExecuteConnector:        MakeExecuteConnectorEndpoint(s),
	}
}

func MakeListConnectorsEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, _ any) (any, error) {
		res, err := s.ListConnectors(ctx)
		return ListConnectorsResponse{Connectors: res, Err: err}, nil
	}
}

func MakeCreateConnectorEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(CreateConnectorRequest)
		res, err := s.CreateConnector(ctx, req.Connector)
		return CreateConnectorResponse{Connector: res, Err: err}, nil
	}
}

func MakeUpdateConnectorEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(UpdateConnectorRequest)
		err := s.UpdateConnector(ctx, req.Connector)
		return UpdateConnectorResponse{Err: err}, nil
	}
}

func MakeDeleteConnectorEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(DeleteConnectorRequest)
		id, err := uuid.Parse(req.ID)
		if err != nil {
			return DeleteConnectorResponse{Err: err}, nil
		}
		err = s.DeleteConnector(ctx, id)
		return DeleteConnectorResponse{Err: err}, nil
	}
}

func MakeListConnectorInstancesEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ListConnectorInstancesRequest)
		projectID, err := uuid.Parse(req.ProjectID)
		if err != nil {
			return ListConnectorInstancesResponse{Err: err}, nil
		}
		res, err := s.ListConnectorInstances(ctx, projectID)
		return ListConnectorInstancesResponse{Instances: res, Err: err}, nil
	}
}

func MakeCreateConnectorInstanceEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(CreateConnectorInstanceRequest)
		res, err := s.CreateConnectorInstance(ctx, req.Instance)
		return CreateConnectorInstanceResponse{Instance: res, Err: err}, nil
	}
}

func MakeUpdateConnectorInstanceEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(UpdateConnectorInstanceRequest)
		err := s.UpdateConnectorInstance(ctx, req.Instance)
		return UpdateConnectorInstanceResponse{Err: err}, nil
	}
}

func MakeDeleteConnectorInstanceEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(DeleteConnectorInstanceRequest)
		id, err := uuid.Parse(req.ID)
		if err != nil {
			return DeleteConnectorInstanceResponse{Err: err}, nil
		}
		err = s.DeleteConnectorInstance(ctx, id)
		return DeleteConnectorInstanceResponse{Err: err}, nil
	}
}

func MakeExecuteConnectorEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ExecuteConnectorRequest)
		res, err := s.ExecuteConnector(ctx, req.ConnectorKey, req.Config, req.Payload)
		return ExecuteConnectorResponse{Result: res, Err: err}, nil
	}
}
