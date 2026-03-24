package signals

import (
	"context"

	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/gsoultan/gobpm/api/proto/endpoints"
	"github.com/gsoultan/gobpm/api/proto/services"
	"github.com/gsoultan/gobpm/server/endpoints/process"
	"github.com/gsoultan/gobpm/server/transports/grpcs/common"
)

type Server struct {
	services.UnimplementedSignalServiceServer
	broadcastSignal grpctransport.Handler
	sendMessage     grpctransport.Handler
}

func NewServer(eps process.Endpoints) *Server {
	return &Server{
		broadcastSignal: grpctransport.NewServer(
			eps.BroadcastSignal,
			decodeGRPCBroadcastSignalRequest,
			encodeGRPCBroadcastSignalResponse,
		),
		sendMessage: grpctransport.NewServer(
			eps.SendMessage,
			decodeGRPCSendMessageRequest,
			encodeGRPCSendMessageResponse,
		),
	}
}

func (s *Server) BroadcastSignal(ctx context.Context, req *endpoints.BroadcastSignalRequest) (*endpoints.BroadcastSignalResponse, error) {
	_, resp, err := s.broadcastSignal.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.BroadcastSignalResponse), nil
}

func (s *Server) SendMessage(ctx context.Context, req *endpoints.SendMessageRequest) (*endpoints.SendMessageResponse, error) {
	_, resp, err := s.sendMessage.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*endpoints.SendMessageResponse), nil
}

func decodeGRPCBroadcastSignalRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.BroadcastSignalRequest)
	return process.BroadcastSignalRequest{
		ProjectID:  req.ProjectId,
		SignalName: req.SignalName,
		Variables:  common.DecodeStruct(req.Variables),
	}, nil
}

func encodeGRPCBroadcastSignalResponse(_ context.Context, response any) (any, error) {
	resp := response.(process.BroadcastSignalResponse)
	return &endpoints.BroadcastSignalResponse{
		Error: common.ErrString(resp.Err),
	}, nil
}

func decodeGRPCSendMessageRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*endpoints.SendMessageRequest)
	return process.SendMessageRequest{
		ProjectID:      req.ProjectId,
		MessageName:    req.MessageName,
		CorrelationKey: req.CorrelationKey,
		Variables:      common.DecodeStruct(req.Variables),
	}, nil
}

func encodeGRPCSendMessageResponse(_ context.Context, response any) (any, error) {
	resp := response.(process.SendMessageResponse)
	return &endpoints.SendMessageResponse{
		Error: common.ErrString(resp.Err),
	}, nil
}
