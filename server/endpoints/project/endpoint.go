package project

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/services"
)

type Endpoints struct {
	CreateProject endpoint.Endpoint
	GetProject    endpoint.Endpoint
	ListProjects  endpoint.Endpoint
	UpdateProject endpoint.Endpoint
	DeleteProject endpoint.Endpoint
}

func MakeEndpoints(s services.ServiceFacade) Endpoints {
	return Endpoints{
		CreateProject: MakeCreateProjectEndpoint(s),
		GetProject:    MakeGetProjectEndpoint(s),
		ListProjects:  MakeListProjectsEndpoint(s),
		UpdateProject: MakeUpdateProjectEndpoint(s),
		DeleteProject: MakeDeleteProjectEndpoint(s),
	}
}

func MakeCreateProjectEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(CreateProjectRequest)
		orgID, err := uuid.Parse(req.OrganizationID)
		if err != nil {
			return CreateProjectResponse{Err: err}, nil
		}
		p, err := s.CreateProject(ctx, orgID, req.Name, req.Description)
		return CreateProjectResponse{Project: p, Err: err}, nil
	}
}

func MakeGetProjectEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(GetProjectRequest)
		id, err := uuid.Parse(req.ID)
		if err != nil {
			return GetProjectResponse{Err: err}, nil
		}
		p, err := s.GetProject(ctx, id)
		return GetProjectResponse{Project: p, Err: err}, nil
	}
}

func MakeListProjectsEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ListProjectsRequest)
		orgID, _ := uuid.Parse(req.OrganizationID)
		projects, err := s.ListProjects(ctx, orgID)
		return ListProjectsResponse{Projects: projects, Err: err}, nil
	}
}

func MakeUpdateProjectEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(UpdateProjectRequest)
		id, err := uuid.Parse(req.ID)
		if err != nil {
			return UpdateProjectResponse{Err: err}, nil
		}
		orgID, _ := uuid.Parse(req.OrganizationID) // optional or could be Nil
		err = s.UpdateProject(ctx, id, orgID, req.Name, req.Description)
		return UpdateProjectResponse{Err: err}, nil
	}
}

func MakeDeleteProjectEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(DeleteProjectRequest)
		id, err := uuid.Parse(req.ID)
		if err != nil {
			return DeleteProjectResponse{Err: err}, nil
		}
		err = s.DeleteProject(ctx, id)
		return DeleteProjectResponse{Err: err}, nil
	}
}
