package external_task

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/gsoultan/gobpm/server/domains/services"
)

type Endpoints struct {
	FetchAndLockExternal  endpoint.Endpoint
	CompleteExternal      endpoint.Endpoint
	HandleExternalFailure endpoint.Endpoint
}

func MakeEndpoints(s services.ServiceFacade) Endpoints {
	return Endpoints{
		FetchAndLockExternal:  MakeFetchAndLockExternalEndpoint(s),
		CompleteExternal:      MakeCompleteExternalEndpoint(s),
		HandleExternalFailure: MakeHandleExternalFailureEndpoint(s),
	}
}

func MakeFetchAndLockExternalEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(FetchAndLockExternalRequest)
		tasks, err := s.FetchAndLock(ctx, req.Topic, req.WorkerID, req.MaxTasks, req.LockDuration)
		if err != nil {
			return FetchAndLockExternalResponse{Error: err.Error()}, nil
		}
		return FetchAndLockExternalResponse{Tasks: tasks}, nil
	}
}

func MakeCompleteExternalEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(CompleteExternalRequest)
		err := s.Complete(ctx, req.TaskID, req.WorkerID, req.Variables)
		if err != nil {
			return CompleteExternalResponse{Error: err.Error()}, nil
		}
		return CompleteExternalResponse{}, nil
	}
}

func MakeHandleExternalFailureEndpoint(s services.ServiceFacade) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(HandleExternalFailureRequest)
		err := s.HandleFailure(ctx, req.TaskID, req.WorkerID, req.ErrorMessage, req.ErrorDetails, req.Retries, req.RetryTimeout)
		if err != nil {
			return HandleExternalFailureResponse{Error: err.Error()}, nil
		}
		return HandleExternalFailureResponse{}, nil
	}
}
