package client

import (
	"context"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/client/commands"
	"github.com/putalexey/goph-keeper/internal/client/config"
	"github.com/putalexey/goph-keeper/internal/client/gproto"
	proto "github.com/putalexey/goph-keeper/internal/common/gproto"
	"go.uber.org/zap"
)

type Client struct {
	proto.GKServerClient
	logger   *zap.SugaredLogger
	config   *config.ClientConfig
	Close    func()
	Commands []commands.Command
}

func NewClient(ctx context.Context, logger *zap.SugaredLogger, config *config.ClientConfig) (*Client, error) {
	c, closeFn, err := gproto.NewGopherGRPCClient(ctx, logger, config)
	if err != nil {
		return nil, err
	}
	app := &Client{
		GKServerClient: c,
		logger:         logger,
		config:         config,
		Close:          closeFn,
	}
	app.Commands = []commands.Command{
		commands.NewPingCommand(logger, app),
	}
	return app, nil
}

func (c *Client) ProcessCommand(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Println("enter command")
		return
	}
	for _, command := range c.Commands {
		if command.GetName() == args[0] {
			err := command.Handle(ctx, args[1:])
			if err != nil {
				fmt.Println(err.Error())
				c.logger.Error(err)
			}
			return
		}
	}
	//if args[0] == "ping" {
	//	response, err := c.Ping(context.Background(), &proto.PingPong{Message: "ping"})
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		c.logger.Error(err)
	//		return
	//	}
	//	fmt.Println(response.Message)
	//	return
	//}
	fmt.Println("Unknown command: ", args[0])
}
