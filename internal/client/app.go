package client

import (
	"context"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/client/commands"
	"github.com/putalexey/goph-keeper/internal/client/config"
	"github.com/putalexey/goph-keeper/internal/client/gproto"
	"github.com/putalexey/goph-keeper/internal/client/storage"
	proto "github.com/putalexey/goph-keeper/internal/common/gproto"
	"go.uber.org/zap"
)

type Client struct {
	proto.GKServerClient
	logger   *zap.SugaredLogger
	config   *config.ClientConfig
	Close    func()
	Commands []commands.Command
	Storage  storage.Storager
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
		Storage:        storage.NewJSONStorage(config.StoragePath),
	}
	app.Commands = []commands.Command{
		commands.NewPingCommand(logger, app),
		commands.NewRegisterCommand(logger, app, app.Storage),
	}
	return app, nil
}

func (c *Client) ProcessCommand(ctx context.Context, args []string) {
	err := c.beforeCommand()
	if err != nil {
		fmt.Println(err.Error())
		c.logger.Error(err)
		return
	}
	defer func() {
		err := c.afterCommand()
		if err != nil {
			fmt.Println(err.Error())
			c.logger.Error(err)
		}
	}()
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
	fmt.Println("Unknown command: ", args[0])
}

func (c *Client) beforeCommand() error {
	err := c.Storage.Load()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) afterCommand() error {
	err := c.Storage.Save()
	if err != nil {
		return err
	}
	return nil
}
