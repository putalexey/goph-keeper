package interfaces

import (
	"context"
	"github.com/pkg/errors"
	"github.com/putalexey/goph-keeper/internal/common/gproto"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"github.com/putalexey/goph-keeper/internal/common/utils/password"
	"github.com/putalexey/goph-keeper/internal/server/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"strings"
)

func NewGopherGRPCServer(ctx context.Context, logger *zap.SugaredLogger, storages *storage.StoragesContainer, address string) *GopherGRPCServer {
	s := grpc.NewServer()
	gopherGRPC := GopherGRPCServer{
		grpcServer: s,
		ctx:        ctx,
		storages:   storages,
		address:    address,
		logger:     logger,
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
	logger     *zap.SugaredLogger
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

func (s *GopherGRPCServer) Ping(_ context.Context, _ *gproto.PingPong) (*gproto.PingPong, error) {
	return &gproto.PingPong{Message: "pong"}, nil
}

func (s *GopherGRPCServer) Register(ctx context.Context, request *gproto.RegisterRequest) (*gproto.RegisterResponse, error) {
	var err error

	login := strings.TrimSpace(request.Login)
	pass := strings.TrimSpace(request.Password)
	passHash, err := password.PasswordHash(pass)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = s.storages.UserStorage.FindByLogin(ctx, login)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "user already registered")
	} else if !errors.Is(err, storage.ErrNotFound) {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	user := &models.User{
		Login:    login,
		Password: passHash,
	}
	err = s.storages.UserStorage.Create(ctx, user)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	auth, err := s.storages.AuthStorage.GenerateForUser(ctx, user)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
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
	var err error

	login := strings.TrimSpace(request.Login)
	pass := strings.TrimSpace(request.Password)
	if len(login) == 0 || len(pass) == 0 {
		return nil, status.Error(codes.InvalidArgument, "login or password is empty")
	}

	user, err := s.storages.UserStorage.FindByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "login or password is not correct")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !password.PasswordCheck(pass, user.Password) {
		p, _ := peer.FromContext(ctx)
		clientIp := p.Addr.String()

		log.Printf("failed login attempt: (%s:%s)\n", clientIp, login)
		return nil, status.Error(codes.NotFound, "login or password is not correct")
	}

	auth, err := s.storages.AuthStorage.GenerateForUser(ctx, user)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &gproto.AuthorizeResponse{
		AuthToken: auth.Token,
		User: &gproto.User{
			UUID:  user.UUID,
			Login: user.Login,
		},
	}, nil
}

func (s *GopherGRPCServer) CreateRecord(ctx context.Context, request *gproto.CreateRecordRequest) (*gproto.CreateRecordResponse, error) {
	user, err := s.authenticate(ctx, request.AuthToken)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(request.Name)
	typeName := strings.TrimSpace(request.Type)
	comment := strings.TrimSpace(request.Comment)
	_, err = s.storages.RecordStorage.GetByUserUUIDAndName(ctx, user.UUID, name)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "record with that name already exists")
	} else if !errors.Is(err, storage.ErrNotFound) {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	record := &models.Record{
		UserUUID: user.UUID,
		Name:     name,
		Type:     typeName,
		Data:     request.Data,
		Comment:  comment,
	}
	err = s.storages.RecordStorage.Create(ctx, record)
	if err != nil {
		return nil, err
	}

	return &gproto.CreateRecordResponse{Record: recordFromModel(record)}, nil
}

func (s *GopherGRPCServer) UpdateRecordField(ctx context.Context, request *gproto.UpdateRecordFieldsRequest) (*gproto.UpdateRecordFieldsResponse, error) {
	user, err := s.authenticate(ctx, request.AuthToken)
	if err != nil {
		return nil, err
	}

	record, err := s.storages.RecordStorage.GetByUUID(ctx, request.UUID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "record not found")
		}
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, "server error")
	}
	if record.UserUUID != user.UUID {
		return nil, status.Error(codes.NotFound, "record not found")
	}

	switch models.RecordField(request.Field) {
	case models.RecordFieldName:
		record.Name = strings.TrimSpace(string(request.Value))
	case models.RecordFieldData:
		record.Data = request.Value
	case models.RecordFieldComment:
		record.Comment = strings.TrimSpace(string(request.Value))
	}

	err = s.storages.RecordStorage.Update(ctx, record)
	if err != nil {
		return nil, err
	}

	return &gproto.UpdateRecordFieldsResponse{Record: recordFromModel(record)}, nil
}

func (s *GopherGRPCServer) DeleteRecord(ctx context.Context, request *gproto.DeleteRecordRequest) (*gproto.Empty, error) {
	user, err := s.authenticate(ctx, request.AuthToken)
	if err != nil {
		return nil, err
	}

	record, err := s.storages.RecordStorage.GetByUUID(ctx, request.UUID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "record not found")
		}
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, "server error")
	}
	if record.UserUUID != user.UUID {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "record not found")
		}
	}

	err = s.storages.RecordStorage.Delete(ctx, record)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "record not found")
		}
		return nil, err
	}

	return &gproto.Empty{}, nil
}

func (s *GopherGRPCServer) GetRecord(ctx context.Context, request *gproto.GetRecordRequest) (*gproto.GetRecordResponse, error) {
	user, err := s.authenticate(ctx, request.AuthToken)
	if err != nil {
		return nil, err
	}

	record, err := s.storages.RecordStorage.GetByUserUUIDAndName(ctx, user.UUID, request.Name)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "record not found")
		}
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, "server error")
	}
	if record.UserUUID != user.UUID {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "record not found")
		}
	}

	return &gproto.GetRecordResponse{Record: recordFromModel(record)}, nil
}

func (s *GopherGRPCServer) GetRecords(ctx context.Context, request *gproto.GetRecordsRequest) (*gproto.GetRecordsResponse, error) {
	user, err := s.authenticate(ctx, request.AuthToken)
	if err != nil {
		return nil, err
	}

	records, err := s.storages.RecordStorage.FindByUserUUID(ctx, user.UUID)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, "server error")
	}
	list := make([]*gproto.RecordListItem, 0, len(records))
	for _, record := range records {
		var createdAt *timestamppb.Timestamp
		if record.CreatedAt != nil {
			createdAt = timestamppb.New(*record.CreatedAt)
		}
		var updatedAt *timestamppb.Timestamp
		if record.UpdatedAt != nil {
			updatedAt = timestamppb.New(*record.UpdatedAt)
		}
		list = append(list, &gproto.RecordListItem{
			UUID:      record.UUID,
			Name:      record.Name,
			Type:      record.Type,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	return &gproto.GetRecordsResponse{Records: list}, nil
}

//func (s *GopherGRPCServer) GetUpdates(ctx context.Context, request *gproto.GetUpdatesRequest) (*gproto.GetUpdatesResponse, error) {
//	//return nil, status.Errorf(codes.Unimplemented, "method GetUpdates not implemented")
//	return s.UnimplementedGKServerServer.GetUpdates(ctx, request)
//}

func (s *GopherGRPCServer) authenticate(ctx context.Context, token string) (*models.User, error) {
	if len(token) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no authentication provided")
	}
	auth, err := s.storages.AuthStorage.FindByToken(ctx, token)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.Unauthenticated, "not authorized")
		}
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, "server error")
	}
	return s.getUser(ctx, auth)
}

func (s *GopherGRPCServer) getUser(ctx context.Context, auth *models.Auth) (*models.User, error) {
	user, err := s.storages.UserStorage.FindByUUID(ctx, auth.UserUUID)
	return user, err
}

func recordFromModel(record *models.Record) *gproto.Record {
	if record == nil {
		return nil
	}
	grecord := &gproto.Record{
		UUID:    record.UUID,
		Name:    record.Name,
		Type:    record.Type,
		Data:    record.Data,
		Comment: record.Comment,
	}

	if record.CreatedAt != nil {
		grecord.CreatedAt = timestamppb.New(*record.CreatedAt)
	}
	if record.UpdatedAt != nil {
		grecord.UpdatedAt = timestamppb.New(*record.UpdatedAt)
	}
	if record.DeletedAt != nil {
		grecord.DeletedAt = timestamppb.New(*record.DeletedAt)
	}

	return grecord
}
