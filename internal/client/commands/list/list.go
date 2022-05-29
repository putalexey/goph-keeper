package list

import (
	"bufio"
	"context"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/client/commands"
	"github.com/putalexey/goph-keeper/internal/client/storage"
	"github.com/putalexey/goph-keeper/internal/common/gproto"
	"go.uber.org/zap"
	"os"
)

type List struct {
	logger  *zap.SugaredLogger
	remote  gproto.GKServerClient
	storage storage.Storager
}

func NewListCommand(logger *zap.SugaredLogger, remote gproto.GKServerClient, storage storage.Storager) *List {
	return &List{logger: logger, remote: remote, storage: storage}
}

func (c *List) GetName() string {
	return "list"
}

func (c *List) GetHelp() string {
	return `list records syntax:
gk-client list`
}

func (c *List) Handle(ctx context.Context, args []string) error {
	if len(c.storage.GetToken()) == 0 {
		return commands.ErrNotAuthorized
	}

	response, err := c.remote.GetRecords(ctx, &gproto.GetRecordsRequest{
		AuthToken: c.storage.GetToken(),
	})
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	err = printRecords(response.Records, args, reader)
	if err != nil {
		return err
	}

	return nil
}

func printRecords(records []*gproto.RecordListItem, _ []string, _ *bufio.Reader) error {
	fmt.Println("<Name>\t<Type>")
	for _, record := range records {
		fmt.Printf("%s\t%s\n", record.Name, record.Type)
	}
	return nil
}
