package projects

import (
	"context"

	"connectrpc.com/connect"
	pbendpoints "github.com/gsoultan/gobpm/api/proto/endpoints"
	pbentities "github.com/gsoultan/gobpm/api/proto/entities"
	"github.com/gsoultan/gobpm/server/endpoints/project"
	"github.com/gsoultan/gobpm/server/transports/adapters"
	"github.com/gsoultan/gobpm/server/transports/grpcs/common"
)

type ProjectHandler struct {
	eps project.Endpoints
}

func NewHandler(eps project.Endpoints) *ProjectHandler {
	return &ProjectHandler{eps: eps}
}

func (h *ProjectHandler) CreateProject(ctx context.Context, req *connect.Request[pbendpoints.CreateProjectRequest]) (*connect.Response[pbendpoints.CreateProjectResponse], error) {
	response, err := h.eps.CreateProject(ctx, project.CreateProjectRequest{
		OrganizationID: req.Msg.OrganizationId,
		Name:           req.Msg.Name,
		Description:    req.Msg.Description,
	})
	if err != nil {
		return nil, err
	}
	resp := response.(project.CreateProjectResponse)
	return connect.NewResponse(&pbendpoints.CreateProjectResponse{
		Project: adapters.ProjectPBAdapter{Project: resp.Project}.ToProto(),
		Error:   common.ErrString(resp.Err),
	}), nil
}

func (h *ProjectHandler) GetProject(ctx context.Context, req *connect.Request[pbendpoints.GetProjectRequest]) (*connect.Response[pbendpoints.GetProjectResponse], error) {
	response, err := h.eps.GetProject(ctx, project.GetProjectRequest{
		ID: req.Msg.Id,
	})
	if err != nil {
		return nil, err
	}
	resp := response.(project.GetProjectResponse)
	return connect.NewResponse(&pbendpoints.GetProjectResponse{
		Project: adapters.ProjectPBAdapter{Project: resp.Project}.ToProto(),
		Error:   common.ErrString(resp.Err),
	}), nil
}

func (h *ProjectHandler) ListProjects(ctx context.Context, req *connect.Request[pbendpoints.ListProjectsRequest]) (*connect.Response[pbendpoints.ListProjectsResponse], error) {
	response, err := h.eps.ListProjects(ctx, project.ListProjectsRequest{
		OrganizationID: req.Msg.OrganizationId,
	})
	if err != nil {
		return nil, err
	}
	resp := response.(project.ListProjectsResponse)
	pbProjects := make([]*pbentities.Project, len(resp.Projects))
	for i, p := range resp.Projects {
		pbProjects[i] = adapters.ProjectPBAdapter{Project: p}.ToProto()
	}
	return connect.NewResponse(&pbendpoints.ListProjectsResponse{
		Projects: pbProjects,
		Error:    common.ErrString(resp.Err),
	}), nil
}

func (h *ProjectHandler) UpdateProject(ctx context.Context, req *connect.Request[pbendpoints.UpdateProjectRequest]) (*connect.Response[pbendpoints.UpdateProjectResponse], error) {
	response, err := h.eps.UpdateProject(ctx, project.UpdateProjectRequest{
		ID:             req.Msg.Id,
		OrganizationID: req.Msg.OrganizationId,
		Name:           req.Msg.Name,
		Description:    req.Msg.Description,
	})
	if err != nil {
		return nil, err
	}
	resp := response.(project.UpdateProjectResponse)
	return connect.NewResponse(&pbendpoints.UpdateProjectResponse{
		Error: common.ErrString(resp.Err),
	}), nil
}

func (h *ProjectHandler) DeleteProject(ctx context.Context, req *connect.Request[pbendpoints.DeleteProjectRequest]) (*connect.Response[pbendpoints.DeleteProjectResponse], error) {
	response, err := h.eps.DeleteProject(ctx, project.DeleteProjectRequest{
		ID: req.Msg.Id,
	})
	if err != nil {
		return nil, err
	}
	resp := response.(project.DeleteProjectResponse)
	return connect.NewResponse(&pbendpoints.DeleteProjectResponse{
		Error: common.ErrString(resp.Err),
	}), nil
}
