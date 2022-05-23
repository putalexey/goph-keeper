package interfaces

import (
	"context"
	"github.com/putalexey/goph-keeper/internal/common/gproto"
	"github.com/putalexey/goph-keeper/internal/server/storage"
	"google.golang.org/grpc"
	"net"
)

func NewGopherGRPCServer(
	ctx context.Context,
	storages *storage.StoragesContainer,
	address string,
) *GopherGRPCServer {
	s := grpc.NewServer()
	gopherGRPC := GopherGRPCServer{
		grpcServer: s,
		ctx:        ctx,
		storages:   storages,
		address:    address,
	}
	gproto.RegisterGKServerServer(s, &gopherGRPC)
	return &gopherGRPC
}

type GopherGRPCServer struct {
	gproto.UnimplementedGKServerServer
	ctx        context.Context
	grpcServer *grpc.Server
	storages   *storage.StoragesContainer
	address    string
}

func (s *GopherGRPCServer) Serve() error {
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	go func() {
		<-s.ctx.Done()
		s.grpcServer.GracefulStop()
	}()
	return s.grpcServer.Serve(listen)
}

func (s *GopherGRPCServer) Ping(_ context.Context, ping *gproto.PingPong) (*gproto.PingPong, error) {
	return &gproto.PingPong{Message: "pong"}, nil
}

func (s *GopherGRPCServer) Register(ctx context.Context, empty *gproto.Empty) (*gproto.Empty, error) {
	//return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
	return s.UnimplementedGKServerServer.Register(ctx, empty)
}
func (s *GopherGRPCServer) Authorize(ctx context.Context, empty *gproto.Empty) (*gproto.Empty, error) {
	//return nil, status.Errorf(codes.Unimplemented, "method Authorize not implemented")
	return s.UnimplementedGKServerServer.Authorize(ctx, empty)
}
func (s *GopherGRPCServer) CreateRecord(ctx context.Context, empty *gproto.Empty) (*gproto.Empty, error) {
	//return nil, status.Errorf(codes.Unimplemented, "method CreateRecord not implemented")
	return s.UnimplementedGKServerServer.CreateRecord(ctx, empty)
}
func (s *GopherGRPCServer) UpdateRecord(ctx context.Context, empty *gproto.Empty) (*gproto.Empty, error) {
	//return nil, status.Errorf(codes.Unimplemented, "method UpdateRecord not implemented")
	return s.UnimplementedGKServerServer.UpdateRecord(ctx, empty)
}
func (s *GopherGRPCServer) DeleteRecord(ctx context.Context, empty *gproto.Empty) (*gproto.Empty, error) {
	//return nil, status.Errorf(codes.Unimplemented, "method DeleteRecord not implemented")
	return s.UnimplementedGKServerServer.DeleteRecord(ctx, empty)
}
func (s *GopherGRPCServer) GetUpdates(ctx context.Context, empty *gproto.Empty) (*gproto.Empty, error) {
	//return nil, status.Errorf(codes.Unimplemented, "method GetUpdates not implemented")
	return s.UnimplementedGKServerServer.GetUpdates(ctx, empty)
}
