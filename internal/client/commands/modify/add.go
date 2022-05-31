package modify

import (
	"bufio"
	"context"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/client/commands"
	"github.com/putalexey/goph-keeper/internal/client/storage"
	"github.com/putalexey/goph-keeper/internal/common/gproto"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"go.uber.org/zap"
	"os"
)

type Add struct {
	logger  *zap.SugaredLogger
	remote  gproto.GKServerClient
	storage storage.Storager
}

func NewAddCommand(logger *zap.SugaredLogger, remote gproto.GKServerClient, storage storage.Storager) *Add {
	return &Add{logger: logger, remote: remote, storage: storage}
}

func (c *Add) GetName() string {
	return "add"
}

func (c *Add) GetFullDescription() string {
	return `Usage:
    gk-client add [type]
    gk-client add text [record_name] [text] [comment]
    gk-client add file [record_name] [filepath] [comment]
    gk-client add login [record_name] [login] [password] [comment]
    gk-client add card [record_name]

Saves new record to the server`
}

func (c *Add) GetShortDescription() string {
	return "add new record"
}

func (c *Add) Handle(ctx context.Context, args []string) error {
	var (
		err      error
		name     string
		typeName string
		data     []byte
		comment  string
	)

	if len(c.storage.GetToken()) == 0 {
		return commands.ErrNotAuthorized
	}

	reader := bufio.NewReader(os.Stdin)

	typeName, args, err = readRecordType(reader, args...)
	if err != nil {
		return err
	}

	name, args, err = readRecordName(reader, args...)
	if err != nil {
		return err
	}

	data, args, err = readRecordValue(typeName, reader, args...)
	if err != nil {
		return err
	}

	comment, args, err = readRecordComment(reader, args...)
	if err != nil {
		return err
	}

	fmt.Printf("adding record \"%s\" with type \"%s\"\n", name, typeName)

	_, err = c.remote.CreateRecord(ctx, &gproto.CreateRecordRequest{
		AuthToken: c.storage.GetToken(),
		Name:      name,
		Type:      typeName,
		Data:      data,
		Comment:   comment,
	})
	if err != nil {
		return err
	}

	fmt.Println("Successful added")

	return nil
}

func readRecordValue(typeName string, reader *bufio.Reader, args ...string) ([]byte, []string, error) {
	switch typeName {
	case models.TypeText:
		return readRecordValueText(reader, args...)
	case models.TypeFile:
		return readRecordValueFile(reader, args...)
	case models.TypeLogin:
		return readRecordValueLogin(reader, args...)
	case models.TypeCard:
		return readNewRecordValueBankCard(reader, args...)
	}
	return []byte{}, args, nil
}
