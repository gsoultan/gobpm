package notification

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/services"
)

type Endpoints struct {
	ListNotifications  endpoint.Endpoint
	MarkAsRead         endpoint.Endpoint
	MarkAllAsRead      endpoint.Endpoint
	DeleteNotification endpoint.Endpoint
}

func MakeEndpoints(s services.ServiceFacade) Endpoints {
	return Endpoints{
		ListNotifications:  MakeListNotificationsEndpoint(s),
		MarkAsRead:         MakeMarkAsReadEndpoint(s),
		MarkAllAsRead:      MakeMarkAllAsReadEndpoint(s),
		DeleteNotification: MakeDeleteNotificationEndpoint(s),
	}
}

func MakeListNotificationsEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ListNotificationsRequest)
		ns, err := s.ListByUser(ctx, req.UserID)
		if err != nil {
			return ListNotificationsResponse{Error: err.Error()}, nil
		}
		return ListNotificationsResponse{Notifications: ns}, nil
	}
}

func MakeMarkAsReadEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(MarkAsReadRequest)
		id, err := uuid.Parse(req.ID)
		if err != nil {
			return MarkAsReadResponse{Error: err.Error()}, nil
		}
		err = s.MarkAsRead(ctx, id)
		if err != nil {
			return MarkAsReadResponse{Error: err.Error()}, nil
		}
		return MarkAsReadResponse{}, nil
	}
}

func MakeMarkAllAsReadEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(MarkAllAsReadRequest)
		err := s.MarkAllAsRead(ctx, req.UserID)
		if err != nil {
			return MarkAllAsReadResponse{Error: err.Error()}, nil
		}
		return MarkAllAsReadResponse{}, nil
	}
}

func MakeDeleteNotificationEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(DeleteNotificationRequest)
		id, err := uuid.Parse(req.ID)
		if err != nil {
			return DeleteNotificationResponse{Error: err.Error()}, nil
		}
		err = s.Delete(ctx, id)
		if err != nil {
			return DeleteNotificationResponse{Error: err.Error()}, nil
		}
		return DeleteNotificationResponse{}, nil
	}
}
