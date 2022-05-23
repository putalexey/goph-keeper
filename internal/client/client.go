package client

import (
	"context"
	"github.com/putalexey/goph-keeper/internal/client/config"
	"github.com/putalexey/goph-keeper/internal/client/gproto"
	proto "github.com/putalexey/goph-keeper/internal/common/gproto"
	"go.uber.org/zap"
)

type Client struct {
	logger *zap.SugaredLogger
	config *config.ClientConfig
	remote proto.GKServerClient
}

func NewClient(ctx context.Context, logger *zap.SugaredLogger, config *config.ClientConfig) *Client {
	return &Client{
		logger: logger,
		config: config,
		remote: gproto.NewGopherGRPCClient(ctx, logger, config),
	}
}
