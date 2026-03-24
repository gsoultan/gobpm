package processes

import (
	"context"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/gsoultan/gobpm/api/proto/endpoints"
	"github.com/gsoultan/gobpm/api/proto/entities"
	"github.com/gsoultan/gobpm/api/proto/services"
	"github.com/gsoultan/gobpm/server/endpoints/process"
	"github.com/gsoultan/gobpm/server/transports/adapters"
	"github.com/gsoultan/gobpm/server/transports/grpcs/common"
)

type Server struct {
	services.UnimplementedProcessServiceServer
	startProcess     grpctransport.Handler
	getInstance      grpctransport.Handler
	listInstances    grpctransport.Handler
	getExecutionPath grpctransport.Handler
}

func NewServer(eps process.Endpoints) *Server {
	return &Server{
		startProcess: grpctransport.NewServer(
			eps.StartProcess,
			decodeGRPCStartProcessRequest,
			encodeGRPCStartProcessResponse,
		),
		getInstance: grpctransport.NewServer(
			eps.GetInstance,
			decodeGRPCGetInstanceRequest,
			encodeGRPCGetInstanceResponse,
		),
		listInstances: grpctransport.NewServer(
			eps.ListInstances,
			decodeGRPCListInstancesRequest,
			encodeGRPCListInstancesResponse,
		),
		getExecutionPath: grpctransport.NewServer(
			eps.GetExecutionPath,
			decodeGRPCGetExecutionPathRequest,
			encodeGRPCGetExecutionPathResponse,
		),
	}
}

func (s *Server) StartProcess(ctx context.Context, req *endpoints.StartProcessRequest) (*endpoints.StartProcessResponse, error) {
	_, resp, err := s.startProcess.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.StartProcessResponse), nil
}

func (s *Server) GetInstance(ctx context.Context, req *endpoints.GetInstanceRequest) (*endpoints.GetInstanceResponse, error) {
	_, resp, err := s.getInstance.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.GetInstanceResponse), nil
}

func (s *Server) ListInstances(ctx context.Context, req *endpoints.ListInstancesRequest) (*endpoints.ListInstancesResponse, error) {
	_, resp, err := s.listInstances.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.ListInstancesResponse), nil
}

func (s *Server) GetExecutionPath(ctx context.Context, req *endpoints.GetExecutionPathRequest) (*endpoints.GetExecutionPathResponse, error) {
	_, resp, err := s.getExecutionPath.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.GetExecutionPathResponse), nil
}

func decodeGRPCStartProcessRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.StartProcessRequest)
	vars := make(map[string]any)
	if req.Variables != nil {
		vars = req.Variables.AsMap()
	}
	return process.StartProcessRequest{
		ProjectID:     req.ProjectId,
		DefinitionKey: req.DefinitionKey,
		Variables:     vars,
	}, nil
}

func encodeGRPCStartProcessResponse(_ context.Context, response any) (any, error) {
	resp := response.(process.StartProcessResponse)
	return &endpoints.StartProcessResponse{InstanceId: resp.InstanceID.String(), Error: common.ErrString(resp.Err)}, nil
}

func decodeGRPCGetInstanceRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.GetInstanceRequest)
	return process.GetInstanceRequest{ID: req.Id}, nil
}

func encodeGRPCGetInstanceResponse(_ context.Context, response any) (any, error) {
	resp := response.(process.GetInstanceResponse)
	return &endpoints.GetInstanceResponse{
		Instance: adapters.ProcessInstancePBAdapter{Instance: resp.Instance}.ToProto(),
		Error:    common.ErrString(resp.Err),
	}, nil
}

func decodeGRPCListInstancesRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.ListInstancesRequest)
	return process.ListInstancesRequest{ProjectID: req.ProjectId}, nil
}

func encodeGRPCListInstancesResponse(_ context.Context, response any) (any, error) {
	resp := response.(process.ListInstancesResponse)
	var instances []*entities.ProcessInstance
	for _, inst := range resp.Instances {
		instances = append(instances, adapters.ProcessInstancePBAdapter{Instance: inst}.ToProto())
	}
	return &endpoints.ListInstancesResponse{Instances: instances, Error: common.ErrString(resp.Err)}, nil
}

func decodeGRPCGetExecutionPathRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.GetExecutionPathRequest)
	return process.GetExecutionPathRequest{InstanceID: req.InstanceId}, nil
}

func encodeGRPCGetExecutionPathResponse(_ context.Context, response any) (any, error) {
	resp := response.(process.GetExecutionPathResponse)
	freqs := make(map[string]int32, len(resp.Frequencies))
	for k, v := range resp.Frequencies {
		freqs[k] = int32(v)
	}
	return &endpoints.GetExecutionPathResponse{
		Nodes:           adapters.NodesToProto(resp.Nodes),
		NodeFrequencies: freqs,
		Error:           resp.Error,
	}, nil
}
