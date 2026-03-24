package organizations

import (
	"context"

	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/gsoultan/gobpm/api/proto/endpoints"
	"github.com/gsoultan/gobpm/api/proto/entities"
	"github.com/gsoultan/gobpm/api/proto/services"
	"github.com/gsoultan/gobpm/server/endpoints/organization"
	"github.com/gsoultan/gobpm/server/transports/adapters"
	"github.com/gsoultan/gobpm/server/transports/grpcs/common"
)

type Server struct {
	services.UnimplementedOrganizationServiceServer
	createOrganization grpctransport.Handler
	getOrganization    grpctransport.Handler
	listOrganizations  grpctransport.Handler
	updateOrganization grpctransport.Handler
	deleteOrganization grpctransport.Handler
}

func NewServer(eps organization.Endpoints) *Server {
	return &Server{
		createOrganization: grpctransport.NewServer(
			eps.CreateOrganization,
			decodeGRPCCreateOrganizationRequest,
			encodeGRPCCreateOrganizationResponse,
		),
		getOrganization: grpctransport.NewServer(
			eps.GetOrganization,
			decodeGRPCGetOrganizationRequest,
			encodeGRPCGetOrganizationResponse,
		),
		listOrganizations: grpctransport.NewServer(
			eps.ListOrganizations,
			decodeGRPCListOrganizationsRequest,
			encodeGRPCListOrganizationsResponse,
		),
		updateOrganization: grpctransport.NewServer(
			eps.UpdateOrganization,
			decodeGRPCUpdateOrganizationRequest,
			encodeGRPCUpdateOrganizationResponse,
		),
		deleteOrganization: grpctransport.NewServer(
			eps.DeleteOrganization,
			decodeGRPCDeleteOrganizationRequest,
			encodeGRPCDeleteOrganizationResponse,
		),
	}
}

func (s *Server) CreateOrganization(ctx context.Context, req *endpoints.CreateOrganizationRequest) (*endpoints.CreateOrganizationResponse, error) {
	_, resp, err := s.createOrganization.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.CreateOrganizationResponse), nil
}

func (s *Server) GetOrganization(ctx context.Context, req *endpoints.GetOrganizationRequest) (*endpoints.GetOrganizationResponse, error) {
	_, resp, err := s.getOrganization.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.GetOrganizationResponse), nil
}

func (s *Server) ListOrganizations(ctx context.Context, req *endpoints.ListOrganizationsRequest) (*endpoints.ListOrganizationsResponse, error) {
	_, resp, err := s.listOrganizations.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.ListOrganizationsResponse), nil
}

func (s *Server) UpdateOrganization(ctx context.Context, req *endpoints.UpdateOrganizationRequest) (*endpoints.UpdateOrganizationResponse, error) {
	_, resp, err := s.updateOrganization.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.UpdateOrganizationResponse), nil
}

func (s *Server) DeleteOrganization(ctx context.Context, req *endpoints.DeleteOrganizationRequest) (*endpoints.DeleteOrganizationResponse, error) {
	_, resp, err := s.deleteOrganization.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.DeleteOrganizationResponse), nil
}

func decodeGRPCCreateOrganizationRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.CreateOrganizationRequest)
	return organization.CreateOrganizationRequest{Name: req.Name, Description: req.Description}, nil
}

func encodeGRPCCreateOrganizationResponse(_ context.Context, response any) (any, error) {
	resp := response.(organization.CreateOrganizationResponse)
	return &endpoints.CreateOrganizationResponse{
		Organization: adapters.OrganizationPBAdapter{Organization: resp.Organization}.ToProto(),
		Error:        common.ErrString(resp.Err),
	}, nil
}

func decodeGRPCGetOrganizationRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.GetOrganizationRequest)
	return organization.GetOrganizationRequest{ID: req.Id}, nil
}

func encodeGRPCGetOrganizationResponse(_ context.Context, response any) (any, error) {
	resp := response.(organization.GetOrganizationResponse)
	return &endpoints.GetOrganizationResponse{
		Organization: adapters.OrganizationPBAdapter{Organization: resp.Organization}.ToProto(),
		Error:        common.ErrString(resp.Err),
	}, nil
}

func decodeGRPCListOrganizationsRequest(_ context.Context, _ any) (any, error) {
	return organization.ListOrganizationsRequest{}, nil
}

func encodeGRPCListOrganizationsResponse(_ context.Context, response any) (any, error) {
	resp := response.(organization.ListOrganizationsResponse)
	var orgs []*entities.Organization
	for _, o := range resp.Organizations {
		orgs = append(orgs, adapters.OrganizationPBAdapter{Organization: o}.ToProto())
	}
	return &endpoints.ListOrganizationsResponse{Organizations: orgs, Error: common.ErrString(resp.Err)}, nil
}

func decodeGRPCUpdateOrganizationRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.UpdateOrganizationRequest)
	return organization.UpdateOrganizationRequest{ID: req.Id, Name: req.Name, Description: req.Description}, nil
}

func encodeGRPCUpdateOrganizationResponse(_ context.Context, response any) (any, error) {
	resp := response.(organization.UpdateOrganizationResponse)
	return &endpoints.UpdateOrganizationResponse{Error: common.ErrString(resp.Err)}, nil
}

func decodeGRPCDeleteOrganizationRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.DeleteOrganizationRequest)
	return organization.DeleteOrganizationRequest{ID: req.Id}, nil
}

func encodeGRPCDeleteOrganizationResponse(_ context.Context, response any) (any, error) {
	resp := response.(organization.DeleteOrganizationResponse)
	return &endpoints.DeleteOrganizationResponse{Error: common.ErrString(resp.Err)}, nil
}
