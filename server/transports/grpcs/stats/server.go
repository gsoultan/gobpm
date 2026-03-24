package stats

import (
	"context"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/gsoultan/gobpm/api/proto/endpoints"
	"github.com/gsoultan/gobpm/api/proto/services"
	"github.com/gsoultan/gobpm/server/endpoints/process"
	"github.com/gsoultan/gobpm/server/transports/grpcs/common"
)

type Server struct {
	services.UnimplementedStatsServiceServer
	getStats grpctransport.Handler
}

func NewServer(eps process.Endpoints) *Server {
	return &Server{
		getStats: grpctransport.NewServer(
			eps.GetProcessStatistics,
			decodeGRPCGetStatsRequest,
			encodeGRPCGetStatsResponse,
		),
	}
}

func (s *Server) GetProcessStatistics(ctx context.Context, req *endpoints.GetProcessStatisticsRequest) (*endpoints.GetProcessStatisticsResponse, error) {
	_, resp, err := s.getStats.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.GetProcessStatisticsResponse), nil
}

func decodeGRPCGetStatsRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.GetProcessStatisticsRequest)
	return process.GetProcessStatisticsRequest{ProjectID: req.ProjectId}, nil
}

func encodeGRPCGetStatsResponse(_ context.Context, response any) (any, error) {
	resp := response.(process.GetProcessStatisticsResponse)
	return &endpoints.GetProcessStatisticsResponse{
		ActiveInstances:    int32(resp.ActiveInstances),
		CompletedInstances: int32(resp.CompletedInstances),
		FailedInstances:    int32(resp.FailedInstances),
		TotalTasks:         int32(resp.TotalTasks),
		PendingTasks:       int32(resp.PendingTasks),
		Error:              common.ErrString(resp.Err),
	}, nil
}
