package gproto

import (
	"context"
	"github.com/putalexey/goph-keeper/internal/client/config"
	proto "github.com/putalexey/goph-keeper/internal/common/gproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewGopherGRPCClient(ctx context.Context, logger *zap.SugaredLogger, clientConfig *config.ClientConfig) (proto.GKServerClient, func(), error) {
	conn, err := grpc.DialContext(ctx, clientConfig.ServerHost, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}

	client := proto.NewGKServerClient(conn)
	return client, func() {
		err := conn.Close()
		if err != nil {
			logger.Error(err)
		}
	}, nil
}
