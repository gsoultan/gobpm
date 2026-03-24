package external_tasks

import (
	"context"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/google/uuid"
	pbendpoints "github.com/gsoultan/gobpm/api/proto/endpoints"
	pbentities "github.com/gsoultan/gobpm/api/proto/entities"
	"github.com/gsoultan/gobpm/api/proto/services"
	"github.com/gsoultan/gobpm/server/endpoints/external_task"
	"github.com/gsoultan/gobpm/server/transports/adapters"
)

type Server struct {
	services.UnimplementedExternalTaskServiceServer
	fetchAndLockExternal  grpctransport.Handler
	completeExternal      grpctransport.Handler
	handleExternalFailure grpctransport.Handler
}

func NewServer(eps external_task.Endpoints) *Server {
	return &Server{
		fetchAndLockExternal: grpctransport.NewServer(
			eps.FetchAndLockExternal,
			decodeGRPCFetchAndLockExternalRequest,
			encodeGRPCFetchAndLockExternalResponse,
		),
		completeExternal: grpctransport.NewServer(
			eps.CompleteExternal,
			decodeGRPCCompleteExternalRequest,
			encodeGRPCCompleteExternalResponse,
		),
		handleExternalFailure: grpctransport.NewServer(
			eps.HandleExternalFailure,
			decodeGRPCHandleExternalFailureRequest,
			encodeGRPCHandleExternalFailureResponse,
		),
	}
}

func (s *Server) FetchAndLockExternalTasks(ctx context.Context, req *pbendpoints.FetchAndLockExternalTasksRequest) (*pbendpoints.FetchAndLockExternalTasksResponse, error) {
	_, resp, err := s.fetchAndLockExternal.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pbendpoints.FetchAndLockExternalTasksResponse), nil
}

func (s *Server) CompleteExternalTask(ctx context.Context, req *pbendpoints.CompleteExternalTaskRequest) (*pbendpoints.CompleteExternalTaskResponse, error) {
	_, resp, err := s.completeExternal.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pbendpoints.CompleteExternalTaskResponse), nil
}

func (s *Server) HandleExternalTaskFailure(ctx context.Context, req *pbendpoints.HandleExternalTaskFailureRequest) (*pbendpoints.HandleExternalTaskFailureResponse, error) {
	_, resp, err := s.handleExternalFailure.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pbendpoints.HandleExternalTaskFailureResponse), nil
}

func decodeGRPCFetchAndLockExternalRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*pbendpoints.FetchAndLockExternalTasksRequest)
	return external_task.FetchAndLockExternalRequest{
		Topic:        req.Topic,
		WorkerID:     req.WorkerId,
		MaxTasks:     int(req.MaxTasks),
		LockDuration: req.LockDurationMs,
	}, nil
}

func encodeGRPCFetchAndLockExternalResponse(_ context.Context, response any) (any, error) {
	resp := response.(external_task.FetchAndLockExternalResponse)
	tasks := make([]*pbentities.ExternalTask, len(resp.Tasks))
	for i, t := range resp.Tasks {
		tasks[i] = adapters.ExternalTaskPBAdapter{Task: *t}.ToProto()
	}
	return &pbendpoints.FetchAndLockExternalTasksResponse{
		Tasks: tasks,
		Error: resp.Error,
	}, nil
}

func decodeGRPCCompleteExternalRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*pbendpoints.CompleteExternalTaskRequest)
	id, err := uuid.Parse(req.TaskId)
	if err != nil {
		return nil, err
	}
	return external_task.CompleteExternalRequest{
		TaskID:    id,
		WorkerID:  req.WorkerId,
		Variables: req.Variables.AsMap(),
	}, nil
}

func encodeGRPCCompleteExternalResponse(_ context.Context, response any) (any, error) {
	resp := response.(external_task.CompleteExternalResponse)
	return &pbendpoints.CompleteExternalTaskResponse{
		Error: resp.Error,
	}, nil
}

func decodeGRPCHandleExternalFailureRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*pbendpoints.HandleExternalTaskFailureRequest)
	id, err := uuid.Parse(req.TaskId)
	if err != nil {
		return nil, err
	}
	return external_task.HandleExternalFailureRequest{
		TaskID:       id,
		WorkerID:     req.WorkerId,
		ErrorMessage: req.ErrorMessage,
		ErrorDetails: req.ErrorDetails,
		Retries:      int(req.Retries),
		RetryTimeout: req.RetryTimeoutMs,
	}, nil
}

func encodeGRPCHandleExternalFailureResponse(_ context.Context, response any) (any, error) {
	resp := response.(external_task.HandleExternalFailureResponse)
	return &pbendpoints.HandleExternalTaskFailureResponse{
		Error: resp.Error,
	}, nil
}
