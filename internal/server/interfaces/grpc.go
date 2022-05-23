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
) *GopherGRPCServer {
	s := grpc.NewServer()
	gopherGRPC := GopherGRPCServer{
		grpcServer: s,
		ctx:        ctx,
		storages:   storages,
	}
	gproto.RegisterGKServerServer(s, &gopherGRPC)
	return &gopherGRPC
}

type GopherGRPCServer struct {
	gproto.UnimplementedGKServerServer
	ctx        context.Context
	grpcServer *grpc.Server
	storages   *storage.StoragesContainer
}

func (s *GopherGRPCServer) Serve() error {
	listen, err := net.Listen("tcp", ":3030")
	if err != nil {
		return err
	}

	go func() {
		<-s.ctx.Done()
		s.grpcServer.GracefulStop()
	}()
	return s.grpcServer.Serve(listen)
}
