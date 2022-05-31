package get

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/client/commands"
	"github.com/putalexey/goph-keeper/internal/client/storage"
	"github.com/putalexey/goph-keeper/internal/common/gproto"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"go.uber.org/zap"
	"os"
	"strings"
)

type Get struct {
	logger  *zap.SugaredLogger
	remote  gproto.GKServerClient
	storage storage.Storager
}

func NewGetCommand(logger *zap.SugaredLogger, remote gproto.GKServerClient, storage storage.Storager) *Get {
	return &Get{logger: logger, remote: remote, storage: storage}
}

func (c *Get) GetName() string {
	return "get"
}

func (c *Get) GetFullDescription() string {
	return `Usage:
    gk-client get [record_name]
    gk-client get <file_record_name> [filepath]

Show records. If type of record was file, is will be saved to provided path`
}

func (c *Get) GetShortDescription() string {
	return "register new user and authorize"
}

func (c *Get) Handle(ctx context.Context, args []string) error {
	var (
		err  error
		name string
	)

	if len(c.storage.GetToken()) == 0 {
		return commands.ErrNotAuthorized
	}

	reader := bufio.NewReader(os.Stdin)
	name, args, err = readRecordName(reader, args)
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

	err = printRecord(response.Record, args, reader)
	if err != nil {
		return err
	}

	return nil
}

func printRecord(record *gproto.Record, args []string, reader *bufio.Reader) error {
	switch record.Type {
	case models.TypeText:
		return printRecordText(record, args)
	case models.TypeFile:
		return printRecordFile(record, args, reader)
	case models.TypeLogin:
		return printRecordLogin(record, args)
	case models.TypeCard:
		return printRecordBankCard(record, args)
	}
	return nil
}

func validateName(name string) error {
	//if len(name) == 0 || strings.ContainsRune(name, '.') {
	//	return errors.New("name cannot be empty or contain \".\" symbol")
	//}
	if len(name) == 0 {
		return errors.New("name cannot be empty")
	}
	return nil
}

func readRecordName(reader *bufio.Reader, args []string) (string, []string, error) {
	var (
		err  error
		name string
	)
	if len(args) < 1 {
		for {
			fmt.Print("Enter record name: ")
			name, err = reader.ReadString('\n')
			if err != nil {
				return "", nil, err
			}
			name = strings.TrimSpace(name)
			err = validateName(name)
			if err == nil {
				break
			}
			fmt.Println(err.Error())
		}
	} else {
		name = strings.TrimSpace(args[0])
		if err = validateName(name); err != nil {
			return "", nil, err
		}
		args = args[1:]
	}
	return name, args, nil
}

func printRecordText(record *gproto.Record, _ []string) error {
	text := string(record.Data)
	fmt.Println(text)
	fmt.Println("Comment:", record.Comment)
	return nil
}
func printRecordFile(record *gproto.Record, args []string, reader *bufio.Reader) error {
	fileModel, err := models.DecodeFileDataType(record.Data)
	if err != nil {
		return err
	}

	var filepath string
	if len(args) < 1 {
		fmt.Print("Where to save file? ")
		filepath, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
	} else {
		filepath = args[0]
		args = args[1:]
	}

	if isDir(filepath) {
		filepath = strings.TrimRight(filepath, string(os.PathSeparator)) + fileModel.Filename
	}

	err = os.WriteFile(filepath, fileModel.Contents, 0664)
	if err != nil {
		return err
	}
	fmt.Println("Written to:", filepath)
	fmt.Println("Comment:", record.Comment)
	return nil
}
func printRecordLogin(record *gproto.Record, _ []string) error {
	loginPasswordModel, err := models.DecodeLoginDataType(record.Data)
	if err != nil {
		return err
	}
	fmt.Println("Login:", loginPasswordModel.Login)
	fmt.Println("Password:", loginPasswordModel.Password)
	fmt.Println("Comment:", record.Comment)
	return nil
}
func printRecordBankCard(record *gproto.Record, _ []string) error {
	bankCardModel, err := models.DecodeBankCardDataType(record.Data)
	if err != nil {
		return err
	}
	fmt.Println("Number:", bankCardModel.Number)
	if len(bankCardModel.Holder) > 0 {
		fmt.Println("Holder:", bankCardModel.Holder)
	}
	fmt.Println("Expiry:", bankCardModel.ExpMonth, '/', bankCardModel.ExpYear)
	fmt.Println("CVV:", bankCardModel.CVV)
	fmt.Println("Comment:", record.Comment)
	return nil
}

// exists returns whether the given file or directory exists
func isDir(path string) bool {
	finfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return finfo.IsDir()
}
