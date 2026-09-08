package platformuser

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
	ListAccounts  endpoint.Endpoint
	SaveAccount   endpoint.Endpoint
	DeleteAccount endpoint.Endpoint
	SetRoles      endpoint.Endpoint
}

func MakeEndpoints(s services.ServiceFacade) Endpoints {
	return Endpoints{
		ListAccounts:  MakeListAccountsEndpoint(s),
		SaveAccount:   MakeSaveAccountEndpoint(s),
		DeleteAccount: MakeDeleteAccountEndpoint(s),
		SetRoles:      MakeSetRolesEndpoint(s),
	}
}

// MakeListAccountsEndpoint returns the accounts and the roles together.
//
// One call rather than two because the page cannot render either alone: an
// account's roles are names, and the list of names a role could be is the other
// half of the same screen.
func MakeListAccountsEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, _ any) (any, error) {
		accounts, err := s.ListPlatformUsers(ctx)
		if err != nil {
			return ListAccountsResponse{Err: err}, nil
		}
		roles, err := s.ListPlatformRoles(ctx)
		return ListAccountsResponse{Accounts: accounts, Roles: roles, Err: err}, nil
	}
}

func MakeSaveAccountEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(SaveAccountRequest)
		if !ok {
			return nil, fmt.Errorf("platformuser: expected a SaveAccountRequest, got %T", request)
		}
		account := entities.PlatformUser{
			Username:    req.Username,
			FullName:    req.FullName,
			DisplayName: req.DisplayName,
			Email:       req.Email,
			Roles:       req.Roles,
		}

		if req.ID == "" {
			id, err := s.CreatePlatformUser(ctx, account, req.Password)
			if err != nil {
				return SaveAccountResponse{Err: err}, nil
			}
			return SaveAccountResponse{ID: id.String()}, nil
		}

		id, err := uuid.Parse(req.ID)
		if err != nil {
			return SaveAccountResponse{Err: apierr.Invalidf("id %q is not a valid identifier: %v", req.ID, err)}, nil
		}
		account.ID = id
		if err := s.UpdatePlatformUser(ctx, account); err != nil {
			return SaveAccountResponse{Err: err}, nil
		}
		return SaveAccountResponse{ID: req.ID}, nil
	}
}

func MakeDeleteAccountEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(DeleteAccountRequest)
		if !ok {
			return nil, fmt.Errorf("platformuser: expected a DeleteAccountRequest, got %T", request)
		}
		id, err := uuid.Parse(req.ID)
		if err != nil {
			return DeleteAccountResponse{Err: apierr.Invalidf("id %q is not a valid identifier: %v", req.ID, err)}, nil
		}
		return DeleteAccountResponse{Err: s.DeletePlatformUser(ctx, id)}, nil
	}
}

// MakeSetRolesEndpoint replaces an account's grants.
//
// Separate from saving the profile so the refusal that protects the last
// administrator has one place to happen, and so a client that only wants to
// rename somebody cannot revoke their roles by omitting the field.
func MakeSetRolesEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(SetRolesRequest)
		if !ok {
			return nil, fmt.Errorf("platformuser: expected a SetRolesRequest, got %T", request)
		}
		id, err := uuid.Parse(req.ID)
		if err != nil {
			return SetRolesResponse{Err: apierr.Invalidf("id %q is not a valid identifier: %v", req.ID, err)}, nil
		}
		return SetRolesResponse{Err: s.SetPlatformRoles(ctx, id, req.Roles)}, nil
	}
}
