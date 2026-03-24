package collaboration

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/gsoultan/gobpm/server/domains/services"
)

type Endpoints struct {
	BroadcastCollaboration endpoint.Endpoint
}

func MakeEndpoints(s services.ServiceFacade) Endpoints {
	return Endpoints{
		BroadcastCollaboration: MakeBroadcastCollaborationEndpoint(s),
	}
}

func MakeBroadcastCollaborationEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(BroadcastCollaborationRequest)
		err := s.Broadcast(ctx, req.Event)
		return BroadcastCollaborationResponse{Err: err}, nil
	}
}
