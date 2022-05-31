package modify

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

type Edit struct {
	logger  *zap.SugaredLogger
	remote  gproto.GKServerClient
	storage storage.Storager
}

func NewEditCommand(logger *zap.SugaredLogger, remote gproto.GKServerClient, storage storage.Storager) *Edit {
	return &Edit{logger: logger, remote: remote, storage: storage}
}

func (c *Edit) GetName() string {
	return "edit"
}

func (c *Edit) GetFullDescription() string {
	return `Usage: gk-client edit [record_name [field [value|filepath]]]

Edit field of the record saved earlier`
}

func (c *Edit) GetShortDescription() string {
	return "edit field of the saved record"
}

func (c *Edit) Handle(ctx context.Context, args []string) error {
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

	field, value, args, err := readRecordFieldValue(response.Record, reader, args...)
	if err != nil {
		return err
	}

	fmt.Printf("editing field \"%s\" of record \"%s\"\n", field.String(), value)

	_, err = c.remote.UpdateRecordField(ctx, &gproto.UpdateRecordFieldsRequest{
		AuthToken: c.storage.GetToken(),
		UUID:      response.Record.UUID,
		Field:     gproto.Field(field),
		Value:     value,
	})
	if err != nil {
		return err
	}

	fmt.Println("Successful edited")

	return nil
}

func readRecordFieldValue(record *gproto.Record, reader *bufio.Reader, args ...string) (models.RecordField, []byte, []string, error) {
	switch record.Type {
	case models.TypeText:
		return readRecordFieldAndValueText(record, reader, args...)
	case models.TypeFile:
		return readRecordFieldAndValueFile(record, reader, args...)
	case models.TypeLogin:
		return readRecordFieldAndValueLogin(record, reader, args...)
	case models.TypeCard:
		return readRecordFieldAndValueBankCard(record, reader, args...)
	}
	return 0, nil, args, nil
}

func readRecordFieldAndValueText(_ *gproto.Record, reader *bufio.Reader, args ...string) (models.RecordField, []byte, []string, error) {
	var (
		field string
	)
	if len(args) < 1 {
		fmt.Println(`Choose field to update: 
  1) name
  2) comment
  3) text`)
		line, err := reader.ReadString('\n')
		if err != nil {
			return 0, nil, args, err
		}
		field = strings.TrimSpace(line)
	} else {
		field = args[0]
		args = args[1:]
	}
	switch field {
	case "1", "name":
		name, args, err := readRecordName(reader, args...)
		if err != nil {
			return 0, nil, args, err
		}
		return models.RecordFieldName, []byte(name), args, nil
	case "2", "comment":
		comment, args, err := readRecordComment(reader, args...)
		if err != nil {
			return 0, nil, args, err
		}
		return models.RecordFieldComment, []byte(comment), args, nil
	case "3", "text":
		text, args, err := readRecordValueText(reader, args...)
		if err != nil {
			return 0, nil, args, err
		}
		return models.RecordFieldData, text, args, nil
	default:
		return 0, nil, args, errors.New("unknown field")
	}
}

func readRecordFieldAndValueFile(record *gproto.Record, reader *bufio.Reader, args ...string) (models.RecordField, []byte, []string, error) {
	var (
		field string
	)
	if len(args) < 1 {
		fmt.Println(`Choose field to update: 
  1) name
  2) comment
  3) file`)
		line, err := reader.ReadString('\n')
		if err != nil {
			return 0, nil, args, err
		}
		field = strings.TrimSpace(line)
	} else {
		field = args[0]
		args = args[1:]
	}
	switch field {
	case "1", "name":
		name, args, err := readRecordName(reader, args...)
		if err != nil {
			return 0, nil, args, err
		}
		return models.RecordFieldName, []byte(name), args, nil
	case "2", "comment":
		comment, args, err := readRecordComment(reader, args...)
		if err != nil {
			return 0, nil, args, err
		}
		return models.RecordFieldComment, []byte(comment), args, nil
	case "3", "file":
		text, args, err := readRecordValueFile(reader, args...)
		if err != nil {
			return 0, nil, args, err
		}
		return models.RecordFieldData, text, args, nil
	default:
		return 0, nil, args, errors.New("unknown field")
	}
}

func readRecordFieldAndValueLogin(record *gproto.Record, reader *bufio.Reader, args ...string) (models.RecordField, []byte, []string, error) {
	var (
		field string
		model *models.LoginDataType
	)
	model, err := models.DecodeLoginDataType(record.Data)
	if err != nil {
		return 0, nil, args, err
	}

	if len(args) < 1 {
		fmt.Println(`Choose field to update: 
  1) name
  2) comment
  3) login
  4) password`)
		line, err := reader.ReadString('\n')
		if err != nil {
			return 0, nil, args, err
		}
		field = strings.TrimSpace(line)
	} else {
		field = args[0]
		args = args[1:]
	}
	switch field {
	case "1", "name":
		name, args, err := readRecordName(reader, args...)
		if err != nil {
			return 0, nil, args, err
		}
		return models.RecordFieldName, []byte(name), args, nil
	case "2", "comment":
		comment, args, err := readRecordComment(reader, args...)
		if err != nil {
			return 0, nil, args, err
		}
		return models.RecordFieldComment, []byte(comment), args, nil
	case "3", "login":
		login, args, err := readRecordValueLoginLogin(reader, args...)
		if err != nil {
			return 0, nil, args, err
		}
		model.Login = login
	case "4", "password":
		password, args, err := readRecordValueLoginPassword(args...)
		if err != nil {
			return 0, nil, args, err
		}
		model.Password = password
	default:
		return 0, nil, args, errors.New("unknown field")
	}

	data, err := models.EncodeLoginDataType(model)
	if err != nil {
		return 0, nil, args, err
	}
	return models.RecordFieldData, data, args, nil
}

func readRecordFieldAndValueBankCard(record *gproto.Record, reader *bufio.Reader, args ...string) (models.RecordField, []byte, []string, error) {
	var (
		field string
		model *models.BankCardDataType
	)
	model, err := models.DecodeBankCardDataType(record.Data)
	if err != nil {
		return 0, nil, args, err
	}

	if len(args) < 1 {
		fmt.Println(`Choose field to update: 
  1) name
  2) comment
  3) number
  4) holder
  5) expiry
  6) cvv`)
		line, err := reader.ReadString('\n')
		if err != nil {
			return 0, nil, args, err
		}
		field = strings.TrimSpace(line)
	} else {
		field = args[0]
		args = args[1:]
	}
	switch field {
	case "1", "name":
		name, args, err := readRecordName(reader, args...)
		if err != nil {
			return 0, nil, args, err
		}
		return models.RecordFieldName, []byte(name), args, nil
	case "2", "comment":
		comment, args, err := readRecordComment(reader, args...)
		if err != nil {
			return 0, nil, args, err
		}
		return models.RecordFieldComment, []byte(comment), args, nil
	case "3", "number":
		number, err := readCardNumber(reader)
		if err != nil {
			return 0, nil, args, err
		}
		model.Number = number
	case "4", "holder":
		holderName, err := readCardHolderName(reader)
		if err != nil {
			return 0, nil, args, err
		}
		model.Holder = holderName
	case "5", "expiry":
		month, year, err := readCardExpiry(reader)
		if err != nil {
			return 0, nil, args, err
		}
		model.ExpMonth = month
		model.ExpYear = year
	case "6", "cvv":
		cvv, err := readCardCVV(reader)
		if err != nil {
			return 0, nil, args, err
		}
		model.CVV = cvv
	default:
		return 0, nil, args, errors.New("unknown field")
	}

	data, err := models.EncodeBankCardDataType(model)
	if err != nil {
		return 0, nil, args, err
	}
	return models.RecordFieldData, data, args, nil
}
