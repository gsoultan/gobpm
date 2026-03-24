package projects

import (
	"context"

	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/gsoultan/gobpm/api/proto/endpoints"
	"github.com/gsoultan/gobpm/api/proto/entities"
	"github.com/gsoultan/gobpm/api/proto/services"
	"github.com/gsoultan/gobpm/server/endpoints/project"
	"github.com/gsoultan/gobpm/server/transports/adapters"
	"github.com/gsoultan/gobpm/server/transports/grpcs/common"
)

type Server struct {
	services.UnimplementedProjectServiceServer
	createProject grpctransport.Handler
	getProject    grpctransport.Handler
	listProjects  grpctransport.Handler
	updateProject grpctransport.Handler
	deleteProject grpctransport.Handler
}

func NewServer(eps project.Endpoints) *Server {
	return &Server{
		createProject: grpctransport.NewServer(
			eps.CreateProject,
			decodeGRPCCreateProjectRequest,
			encodeGRPCCreateProjectResponse,
		),
		getProject: grpctransport.NewServer(
			eps.GetProject,
			decodeGRPCGetProjectRequest,
			encodeGRPCGetProjectResponse,
		),
		listProjects: grpctransport.NewServer(
			eps.ListProjects,
			decodeGRPCListProjectsRequest,
			encodeGRPCListProjectsResponse,
		),
		updateProject: grpctransport.NewServer(
			eps.UpdateProject,
			decodeGRPCUpdateProjectRequest,
			encodeGRPCUpdateProjectResponse,
		),
		deleteProject: grpctransport.NewServer(
			eps.DeleteProject,
			decodeGRPCDeleteProjectRequest,
			encodeGRPCDeleteProjectResponse,
		),
	}
}

func (s *Server) CreateProject(ctx context.Context, req *endpoints.CreateProjectRequest) (*endpoints.CreateProjectResponse, error) {
	_, resp, err := s.createProject.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.CreateProjectResponse), nil
}

func (s *Server) GetProject(ctx context.Context, req *endpoints.GetProjectRequest) (*endpoints.GetProjectResponse, error) {
	_, resp, err := s.getProject.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.GetProjectResponse), nil
}

func (s *Server) ListProjects(ctx context.Context, req *endpoints.ListProjectsRequest) (*endpoints.ListProjectsResponse, error) {
	_, resp, err := s.listProjects.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.ListProjectsResponse), nil
}

func (s *Server) UpdateProject(ctx context.Context, req *endpoints.UpdateProjectRequest) (*endpoints.UpdateProjectResponse, error) {
	_, resp, err := s.updateProject.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.UpdateProjectResponse), nil
}

func (s *Server) DeleteProject(ctx context.Context, req *endpoints.DeleteProjectRequest) (*endpoints.DeleteProjectResponse, error) {
	_, resp, err := s.deleteProject.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.DeleteProjectResponse), nil
}

func decodeGRPCCreateProjectRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.CreateProjectRequest)
	return project.CreateProjectRequest{
		OrganizationID: req.OrganizationId,
		Name:           req.Name,
		Description:    req.Description,
	}, nil
}

func encodeGRPCCreateProjectResponse(_ context.Context, response any) (any, error) {
	resp := response.(project.CreateProjectResponse)
	return &endpoints.CreateProjectResponse{
		Project: adapters.ProjectPBAdapter{Project: resp.Project}.ToProto(),
		Error:   common.ErrString(resp.Err),
	}, nil
}

func decodeGRPCGetProjectRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.GetProjectRequest)
	return project.GetProjectRequest{ID: req.Id}, nil
}

func encodeGRPCGetProjectResponse(_ context.Context, response any) (any, error) {
	resp := response.(project.GetProjectResponse)
	return &endpoints.GetProjectResponse{
		Project: adapters.ProjectPBAdapter{Project: resp.Project}.ToProto(),
		Error:   common.ErrString(resp.Err),
	}, nil
}

func decodeGRPCListProjectsRequest(_ context.Context, _ any) (any, error) {
	return project.ListProjectsRequest{}, nil
}

func encodeGRPCListProjectsResponse(_ context.Context, response any) (any, error) {
	resp := response.(project.ListProjectsResponse)
	var projects []*entities.Project
	for _, p := range resp.Projects {
		projects = append(projects, adapters.ProjectPBAdapter{Project: p}.ToProto())
	}
	return &endpoints.ListProjectsResponse{Projects: projects, Error: common.ErrString(resp.Err)}, nil
}

func decodeGRPCUpdateProjectRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.UpdateProjectRequest)
	return project.UpdateProjectRequest{
		ID:             req.Id,
		OrganizationID: req.OrganizationId,
		Name:           req.Name,
		Description:    req.Description,
	}, nil
}

func encodeGRPCUpdateProjectResponse(_ context.Context, response any) (any, error) {
	resp := response.(project.UpdateProjectResponse)
	return &endpoints.UpdateProjectResponse{Error: common.ErrString(resp.Err)}, nil
}

func decodeGRPCDeleteProjectRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.DeleteProjectRequest)
	return project.DeleteProjectRequest{ID: req.Id}, nil
}

func encodeGRPCDeleteProjectResponse(_ context.Context, response any) (any, error) {
	resp := response.(project.DeleteProjectResponse)
	return &endpoints.DeleteProjectResponse{Error: common.ErrString(resp.Err)}, nil
}
