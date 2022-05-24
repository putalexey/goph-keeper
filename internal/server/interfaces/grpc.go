package interfaces

import (
	"context"
	"github.com/pkg/errors"
	"github.com/putalexey/goph-keeper/internal/common/gproto"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"github.com/putalexey/goph-keeper/internal/common/utils/password"
	"github.com/putalexey/goph-keeper/internal/server/storage"
	"google.golang.org/grpc"
	"net"
	"strings"
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
func (s *GopherGRPCServer) Register(ctx context.Context, request *gproto.RegisterRequest) (*gproto.RegisterResponse, error) {
	var err error

	login := strings.TrimSpace(request.Login)
	pass := strings.TrimSpace(request.Password)
	passHash, err := password.PasswordHash(pass)
	if err != nil {
		return nil, err
	}

	_, err = s.storages.UserStorage.FindByLogin(ctx, login)
	if err == nil {
		return nil, errors.New("user already registered")
	} else if !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}
	user := &models.User{
		Login:    login,
		Password: passHash,
	}
	err = s.storages.UserStorage.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	auth, err := s.storages.AuthStorage.GenerateForUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &gproto.RegisterResponse{
		AuthToken: auth.Token,
		User: &gproto.User{
			UUID:  user.UUID,
			Login: user.Login,
		},
	}, nil
}
func (s *GopherGRPCServer) Authorize(ctx context.Context, request *gproto.AuthorizeRequest) (*gproto.AuthorizeResponse, error) {
	//return nil, status.Errorf(codes.Unimplemented, "method Authorize not implemented")
	return s.UnimplementedGKServerServer.Authorize(ctx, request)
}
func (s *GopherGRPCServer) CreateRecord(ctx context.Context, request *gproto.CreateRecordRequest) (*gproto.CreateRecordResponse, error) {
	//return nil, status.Errorf(codes.Unimplemented, "method CreateRecord not implemented")
	return s.UnimplementedGKServerServer.CreateRecord(ctx, request)
}
func (s *GopherGRPCServer) UpdateRecord(ctx context.Context, request *gproto.UpdateRecordRequest) (*gproto.UpdateRecordResponse, error) {
	//return nil, status.Errorf(codes.Unimplemented, "method UpdateRecord not implemented")
	return s.UnimplementedGKServerServer.UpdateRecord(ctx, request)
}
func (s *GopherGRPCServer) DeleteRecord(ctx context.Context, request *gproto.DeleteRecordRequest) (*gproto.Empty, error) {
	//return nil, status.Errorf(codes.Unimplemented, "method DeleteRecord not implemented")
	return s.UnimplementedGKServerServer.DeleteRecord(ctx, request)
}
func (s *GopherGRPCServer) GetUpdates(ctx context.Context, request *gproto.GetUpdatesRequest) (*gproto.GetUpdatesResponse, error) {
	//return nil, status.Errorf(codes.Unimplemented, "method GetUpdates not implemented")
	return s.UnimplementedGKServerServer.GetUpdates(ctx, request)
}
