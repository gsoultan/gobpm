package tasks

import (
	"context"

	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/gsoultan/gobpm/api/proto/endpoints"
	"github.com/gsoultan/gobpm/api/proto/entities"
	"github.com/gsoultan/gobpm/api/proto/services"
	"github.com/gsoultan/gobpm/server/endpoints/task"
	"github.com/gsoultan/gobpm/server/transports/adapters"
	"github.com/gsoultan/gobpm/server/transports/grpcs/common"
)

type Server struct {
	services.UnimplementedTaskServiceServer
	getTask               grpctransport.Handler
	listTasks             grpctransport.Handler
	completeTask          grpctransport.Handler
	claimTask             grpctransport.Handler
	unclaimTask           grpctransport.Handler
	listTasksByAssignee   grpctransport.Handler
	listTasksByCandidates grpctransport.Handler
}

func NewServer(eps task.Endpoints) *Server {
	return &Server{
		getTask: grpctransport.NewServer(
			eps.GetTask,
			decodeGRPCGetTaskRequest,
			encodeGRPCGetTaskResponse,
		),
		listTasks: grpctransport.NewServer(
			eps.ListTasks,
			decodeGRPCListTasksRequest,
			encodeGRPCListTasksResponse,
		),
		completeTask: grpctransport.NewServer(
			eps.CompleteTask,
			decodeGRPCCompleteTaskRequest,
			encodeGRPCCompleteTaskResponse,
		),
		claimTask: grpctransport.NewServer(
			eps.ClaimTask,
			decodeGRPCClaimTaskRequest,
			encodeGRPCClaimTaskResponse,
		),
		unclaimTask: grpctransport.NewServer(
			eps.UnclaimTask,
			decodeGRPCUnclaimTaskRequest,
			encodeGRPCUnclaimTaskResponse,
		),
		listTasksByAssignee: grpctransport.NewServer(
			eps.ListTasksByAssignee,
			decodeGRPCListTasksByAssigneeRequest,
			encodeGRPCListTasksResponse,
		),
		listTasksByCandidates: grpctransport.NewServer(
			eps.ListTasksByCandidates,
			decodeGRPCListTasksByCandidatesRequest,
			encodeGRPCListTasksResponse,
		),
	}
}

func (s *Server) GetTask(ctx context.Context, req *endpoints.GetTaskRequest) (*endpoints.GetTaskResponse, error) {
	_, resp, err := s.getTask.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.GetTaskResponse), nil
}

func (s *Server) ListTasks(ctx context.Context, req *endpoints.ListTasksRequest) (*endpoints.ListTasksResponse, error) {
	_, resp, err := s.listTasks.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.ListTasksResponse), nil
}

func (s *Server) CompleteTask(ctx context.Context, req *endpoints.CompleteTaskRequest) (*endpoints.CompleteTaskResponse, error) {
	_, resp, err := s.completeTask.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.CompleteTaskResponse), nil
}

func (s *Server) ClaimTask(ctx context.Context, req *endpoints.ClaimTaskRequest) (*endpoints.ClaimTaskResponse, error) {
	_, resp, err := s.claimTask.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.ClaimTaskResponse), nil
}

func (s *Server) UnclaimTask(ctx context.Context, req *endpoints.UnclaimTaskRequest) (*endpoints.UnclaimTaskResponse, error) {
	_, resp, err := s.unclaimTask.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.UnclaimTaskResponse), nil
}

func (s *Server) ListTasksByAssignee(ctx context.Context, req *endpoints.ListTasksByAssigneeRequest) (*endpoints.ListTasksResponse, error) {
	_, resp, err := s.listTasksByAssignee.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.ListTasksResponse), nil
}

func (s *Server) ListTasksByCandidates(ctx context.Context, req *endpoints.ListTasksByCandidatesRequest) (*endpoints.ListTasksResponse, error) {
	_, resp, err := s.listTasksByCandidates.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.ListTasksResponse), nil
}

func decodeGRPCGetTaskRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.GetTaskRequest)
	return task.GetTaskRequest{ID: req.Id}, nil
}

func encodeGRPCGetTaskResponse(_ context.Context, response any) (any, error) {
	resp := response.(task.GetTaskResponse)
	return &endpoints.GetTaskResponse{
		Task:  adapters.TaskPBAdapter{Task: resp.Task}.ToProto(),
		Error: common.ErrString(resp.Err),
	}, nil
}

func decodeGRPCListTasksRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.ListTasksRequest)
	return task.ListTasksRequest{ProjectID: req.ProjectId}, nil
}

func encodeGRPCListTasksResponse(_ context.Context, response any) (any, error) {
	resp := response.(task.ListTasksResponse)
	var tasks []*entities.Task
	for _, t := range resp.Tasks {
		tasks = append(tasks, adapters.TaskPBAdapter{Task: t}.ToProto())
	}
	return &endpoints.ListTasksResponse{Tasks: tasks, Error: common.ErrString(resp.Err)}, nil
}

func decodeGRPCCompleteTaskRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.CompleteTaskRequest)
	vars := make(map[string]any)
	if req.Variables != nil {
		vars = req.Variables.AsMap()
	}
	return task.CompleteTaskRequest{ID: req.Id, Variables: vars}, nil
}

func encodeGRPCCompleteTaskResponse(_ context.Context, response any) (any, error) {
	resp := response.(task.CompleteTaskResponse)
	return &endpoints.CompleteTaskResponse{Error: common.ErrString(resp.Err)}, nil
}

func decodeGRPCClaimTaskRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.ClaimTaskRequest)
	return task.ClaimTaskRequest{ID: req.Id, UserID: req.UserId}, nil
}

func encodeGRPCClaimTaskResponse(_ context.Context, response any) (any, error) {
	resp := response.(task.CompleteTaskResponse)
	return &endpoints.ClaimTaskResponse{Error: common.ErrString(resp.Err)}, nil
}

func decodeGRPCUnclaimTaskRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.UnclaimTaskRequest)
	return task.UnclaimTaskRequest{ID: req.Id}, nil
}

func encodeGRPCUnclaimTaskResponse(_ context.Context, response any) (any, error) {
	resp := response.(task.CompleteTaskResponse)
	return &endpoints.UnclaimTaskResponse{Error: common.ErrString(resp.Err)}, nil
}

func decodeGRPCListTasksByAssigneeRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.ListTasksByAssigneeRequest)
	return task.ListTasksByAssigneeRequest{Assignee: req.Assignee}, nil
}

func decodeGRPCListTasksByCandidatesRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.ListTasksByCandidatesRequest)
	return task.ListTasksByCandidatesRequest{UserID: req.UserId, Groups: req.Groups}, nil
}
