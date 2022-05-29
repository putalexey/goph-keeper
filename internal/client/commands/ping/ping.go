package ping

import (
	"context"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/common/gproto"
	"go.uber.org/zap"
)

type Ping struct {
	logger *zap.SugaredLogger
	remote gproto.GKServerClient
}

func NewPingCommand(logger *zap.SugaredLogger, remote gproto.GKServerClient) *Ping {
	return &Ping{logger: logger, remote: remote}
}

func (c *Ping) GetName() string {
	return "ping"
}

func (c *Ping) Handle(ctx context.Context, _ []string) error {
	response, err := c.remote.Ping(ctx, &gproto.PingPong{Message: "ping"})
	if err != nil {
		return err
	}
	fmt.Println(response.Message)
	return nil
}
