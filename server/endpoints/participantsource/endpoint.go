package participantsource

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gsoultan/metis/internal/pkg/apierr"
	"github.com/gsoultan/metis/server/domains/entities"
	"github.com/gsoultan/metis/server/domains/services"
)

type Endpoints struct {
	ListSources  endpoint.Endpoint
	SaveSource   endpoint.Endpoint
	DeleteSource endpoint.Endpoint
	SyncSource   endpoint.Endpoint
}

func MakeEndpoints(s services.ServiceFacade) Endpoints {
	return Endpoints{
		ListSources:  MakeListSourcesEndpoint(s),
		SaveSource:   MakeSaveSourceEndpoint(s),
		DeleteSource: MakeDeleteSourceEndpoint(s),
		SyncSource:   MakeSyncSourceEndpoint(s),
	}
}

func MakeListSourcesEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(ListSourcesRequest)
		if !ok {
			return nil, fmt.Errorf("participantsource: expected a ListSourcesRequest, got %T", request)
		}
		projectID, err := uuid.Parse(req.ProjectID)
		if err != nil {
			return ListSourcesResponse{Err: apierr.Invalidf("project_id %q is not a valid identifier: %v", req.ProjectID, err)}, nil
		}
		sources, err := s.ListSources(ctx, projectID)
		return ListSourcesResponse{Sources: sources, Err: err}, nil
	}
}

func MakeSaveSourceEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(SaveSourceRequest)
		if !ok {
			return nil, fmt.Errorf("participantsource: expected a SaveSourceRequest, got %T", request)
		}
		projectID, err := uuid.Parse(req.ProjectID)
		if err != nil {
			return SaveSourceResponse{Err: apierr.Invalidf("project_id %q is not a valid identifier: %v", req.ProjectID, err)}, nil
		}
		source := entities.ParticipantSource{
			Project:   &entities.Project{ID: projectID},
			Name:      req.Name,
			Kind:      req.Kind,
			Config:    req.Config,
			Schedule:  req.Schedule,
			OnMissing: req.OnMissing,
			Enabled:   req.Enabled,
		}
		if req.ID != "" {
			id, err := uuid.Parse(req.ID)
			if err != nil {
				return SaveSourceResponse{Err: apierr.Invalidf("id %q is not a valid identifier: %v", req.ID, err)}, nil
			}
			source.ID = id
		}
		id, err := s.SaveSource(ctx, source)
		return SaveSourceResponse{ID: id, Err: err}, nil
	}
}

func MakeDeleteSourceEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(DeleteSourceRequest)
		if !ok {
			return nil, fmt.Errorf("participantsource: expected a DeleteSourceRequest, got %T", request)
		}
		id, err := uuid.Parse(req.ID)
		if err != nil {
			return DeleteSourceResponse{Err: apierr.Invalidf("id %q is not a valid identifier: %v", req.ID, err)}, nil
		}
		return DeleteSourceResponse{Err: s.DeleteSource(ctx, id)}, nil
	}
}

// MakeSyncSourceEndpoint runs one directory now.
//
// The same path a scheduled run takes, including the lock — so pressing the
// button while a scheduled run is in flight is refused rather than doubling the
// work.
func MakeSyncSourceEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(SyncSourceRequest)
		if !ok {
			return nil, fmt.Errorf("participantsource: expected a SyncSourceRequest, got %T", request)
		}
		id, err := uuid.Parse(req.ID)
		if err != nil {
			return SyncSourceResponse{Err: apierr.Invalidf("id %q is not a valid identifier: %v", req.ID, err)}, nil
		}
		summary, err := s.SyncSource(ctx, id)
		return SyncSourceResponse{
			Created:     summary.Created,
			Updated:     summary.Updated,
			Groups:      summary.Groups,
			Deactivated: summary.Deactivated,
			Problems:    summary.Problems,
			Err:         err,
		}, nil
	}
}
