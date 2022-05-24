package commands

import (
	"context"
	"fmt"
	proto "github.com/putalexey/goph-keeper/internal/common/gproto"
	"go.uber.org/zap"
)

type RegisterCommand struct {
	logger *zap.SugaredLogger
	remote proto.GKServerClient
}

func NewRegisterCommand(logger *zap.SugaredLogger, remote proto.GKServerClient) *RegisterCommand {
	return &RegisterCommand{logger: logger, remote: remote}
}

func (c *RegisterCommand) GetName() string {
	return "register"
}

func (c *RegisterCommand) Handle(ctx context.Context, _ []string) error {
	response, err := c.remote.Register(ctx, &proto.RegisterRequest{
		Login:    "test login",
		Password: "test password",
	})
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", response)
	return nil
}
