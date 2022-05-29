package modify

import (
	"bufio"
	"context"
	"github.com/putalexey/goph-keeper/internal/client/commands"
	"github.com/putalexey/goph-keeper/internal/client/storage"
	"github.com/putalexey/goph-keeper/internal/common/gproto"
	"go.uber.org/zap"
	"os"
)

type Delete struct {
	logger  *zap.SugaredLogger
	remote  gproto.GKServerClient
	storage storage.Storager
}

func NewDeleteCommand(logger *zap.SugaredLogger, remote gproto.GKServerClient, storage storage.Storager) *Delete {
	return &Delete{logger: logger, remote: remote, storage: storage}
}

func (c *Delete) GetName() string {
	return "delete"
}

func (c *Delete) GetHelp() string {
	return `delete record syntax:
gk-client delete
gk-client delete [record_name]`
}

func (c *Delete) Handle(ctx context.Context, args []string) error {
	var (
		err  error
		name string
	)

	if len(c.storage.GetToken()) == 0 {
		return commands.ErrNotAuthorized
	}

	reader := bufio.NewReader(os.Stdin)
	name, args, err = readRecordName(reader, args...)
	if err != nil {
		return err
	}

	response, err := c.remote.GetRecord(ctx, &gproto.GetRecordRequest{
		AuthToken: c.storage.GetToken(),
		Name:      name,
	})
	if err != nil {
		return err
	}

	_, err = c.remote.DeleteRecord(ctx, &gproto.DeleteRecordRequest{
		AuthToken: c.storage.GetToken(),
		UUID:      response.Record.UUID,
	})
	return err
}
