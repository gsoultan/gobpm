package incident

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/services"
)

type Endpoints struct {
	ListIncidents   endpoint.Endpoint
	ResolveIncident endpoint.Endpoint
}

func MakeEndpoints(s services.ServiceFacade) Endpoints {
	return Endpoints{
		ListIncidents:   MakeListIncidentsEndpoint(s),
		ResolveIncident: MakeResolveIncidentEndpoint(s),
	}
}

func MakeListIncidentsEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ListIncidentsRequest)
		id, err := uuid.Parse(req.InstanceID)
		if err != nil {
			return ListIncidentsResponse{Err: err}, nil
		}
		incidents, err := s.ListIncidents(ctx, id)
		return ListIncidentsResponse{Incidents: incidents, Err: err}, nil
	}
}

func MakeResolveIncidentEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ResolveIncidentRequest)
		id, err := uuid.Parse(req.ID)
		if err != nil {
			return ResolveIncidentResponse{Err: err}, nil
		}
		err = s.ResolveIncident(ctx, id)
		return ResolveIncidentResponse{Err: err}, nil
	}
}
