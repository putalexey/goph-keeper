package commands

import (
	"context"
	"fmt"
	proto "github.com/putalexey/goph-keeper/internal/common/gproto"
	"go.uber.org/zap"
)

type PingCommand struct {
	logger *zap.SugaredLogger
	remote proto.GKServerClient
}

func NewPingCommand(logger *zap.SugaredLogger, remote proto.GKServerClient) *PingCommand {
	return &PingCommand{logger: logger, remote: remote}
}

func (c *PingCommand) GetName() string {
	return "ping"
}

func (c *PingCommand) Handle(ctx context.Context, _ []string) error {
	response, err := c.remote.Ping(ctx, &proto.PingPong{Message: "ping"})
	if err != nil {
		return err
	}
	fmt.Println(response.Message)
	return nil
}
