package group

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/gsoultan/gobpm/server/domains/services"
)

type Endpoints struct {
	ListGroups       endpoint.Endpoint
	CreateGroup      endpoint.Endpoint
	GetGroup         endpoint.Endpoint
	UpdateGroup      endpoint.Endpoint
	DeleteGroup      endpoint.Endpoint
	ListGroupMembers endpoint.Endpoint
	AddMembership    endpoint.Endpoint
	RemoveMembership endpoint.Endpoint
	ListUserGroups   endpoint.Endpoint
}

func MakeEndpoints(s services.ServiceFacade) Endpoints {
	return Endpoints{
		ListGroups:       MakeListGroupsEndpoint(s),
		CreateGroup:      MakeCreateGroupEndpoint(s),
		GetGroup:         MakeGetGroupEndpoint(s),
		UpdateGroup:      MakeUpdateGroupEndpoint(s),
		DeleteGroup:      MakeDeleteGroupEndpoint(s),
		ListGroupMembers: MakeListGroupMembersEndpoint(s),
		AddMembership:    MakeAddMembershipEndpoint(s),
		RemoveMembership: MakeRemoveMembershipEndpoint(s),
		ListUserGroups:   MakeListUserGroupsEndpoint(s),
	}
}

func MakeListGroupsEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ListGroupsRequest)
		groups, err := s.ListGroups(ctx, req.OrganizationID)
		return ListGroupsResponse{Groups: groups, Err: err}, nil
	}
}

func MakeCreateGroupEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(CreateGroupRequest)
		err := s.CreateGroup(ctx, req.Group)
		return CreateGroupResponse{Err: err}, nil
	}
}

func MakeGetGroupEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(GetGroupRequest)
		group, err := s.GetGroup(ctx, req.ID)
		return GetGroupResponse{Group: group, Err: err}, nil
	}
}

func MakeUpdateGroupEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(UpdateGroupRequest)
		err := s.UpdateGroup(ctx, req.Group)
		return UpdateGroupResponse{Err: err}, nil
	}
}

func MakeDeleteGroupEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(DeleteGroupRequest)
		err := s.DeleteGroup(ctx, req.ID)
		return DeleteGroupResponse{Err: err}, nil
	}
}

func MakeListGroupMembersEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ListGroupMembersRequest)
		users, err := s.ListGroupMembers(ctx, req.GroupID)
		return ListGroupMembersResponse{Users: users, Err: err}, nil
	}
}

func MakeAddMembershipEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(AddMembershipRequest)
		err := s.AddMembership(ctx, req.UserID, req.GroupID)
		return AddMembershipResponse{Err: err}, nil
	}
}

func MakeRemoveMembershipEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(RemoveMembershipRequest)
		err := s.RemoveMembership(ctx, req.UserID, req.GroupID)
		return RemoveMembershipResponse{Err: err}, nil
	}
}

func MakeListUserGroupsEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ListUserGroupsRequest)
		groups, err := s.ListUserGroups(ctx, req.UserID)
		return ListUserGroupsResponse{Groups: groups, Err: err}, nil
	}
}
