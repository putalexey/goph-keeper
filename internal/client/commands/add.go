package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/client/storage"
	proto "github.com/putalexey/goph-keeper/internal/common/gproto"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"golang.org/x/term"
	"os"
	"path"
	"strconv"
	"strings"
)

type AddCommand struct {
	logger  *zap.SugaredLogger
	remote  proto.GKServerClient
	storage storage.Storager
}

func NewAddCommand(logger *zap.SugaredLogger, remote proto.GKServerClient, storage storage.Storager) *AddCommand {
	return &AddCommand{logger: logger, remote: remote, storage: storage}
}

func (c *AddCommand) GetName() string {
	return "add"
}

func (c *AddCommand) Handle(ctx context.Context, args []string) error {
	var (
		err      error
		name     string
		typeName string
		data     []byte
	)

	reader := bufio.NewReader(os.Stdin)
	if len(args) < 1 {
		name, err = readRecordName(reader)
		if err != nil {
			return err
		}
	} else {
		name = strings.TrimSpace(args[0])
		if err = validateName(name); err != nil {
			return err
		}
	}

	if len(args) < 2 {
		typeName, err = readRecordType(reader)
		if err != nil {
			return err
		}
	} else {
		typeName = strings.TrimSpace(args[1])
		if !slices.Contains(storage.SupportedTypes, typeName) {
			errText := fmt.Sprintf("unknown record type (%s)\nRecord typesn\n", typeName)
			for i, t := range storage.SupportedTypes {
				errText += fmt.Sprintf("  [%d] %s\n", i+1, t)
			}
			return errors.New(errText)
		}
	}

	if len(args) < 3 {
		data, err = readRecordValue(typeName, reader)
	} else {
		data, err = readRecordValue(typeName, reader, args[2:]...)
	}
	if err != nil {
		return err
	}

	fmt.Printf("adding record \"%s\" with type \"%s\" and data (%v)\n", name, typeName, string(data))

	//c.storage.SetToken(response.AuthToken)
	//c.storage.SetLogin(response.User.Login)
	fmt.Println("Successful added")

	return nil
}

func readRecordValue(typeName string, reader *bufio.Reader, args ...string) ([]byte, error) {
	switch typeName {
	case models.TypeText:
		return readRecordValueText(reader, args...)
	case models.TypeFile:
		return readRecordValueFile(reader, args...)
	case models.TypeLogin:
		return readRecordValueLogin(reader, args...)
	case models.TypeBank:
		return readRecordValueBank(reader, args...)
	}
	return []byte{}, nil
}

//readRecordValueText if args passed to this function it returns first element as []byte
//else it asks user to enter text and reads line by line. When two empty line met text considered finished (last two
//empty lines not included to result value)
func readRecordValueText(reader *bufio.Reader, args ...string) ([]byte, error) {
	if len(args) > 0 {
		return []byte(args[0]), nil
	}
	lastBlank := false
	prevLine := ""
	fmt.Println("Enter text (leave two blank lines to finish text)")
	var text = strings.Builder{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if len(strings.TrimSpace(line)) == 0 {
			if lastBlank {
				break
			}
			lastBlank = true
			prevLine = line
		} else {
			if lastBlank {
				text.WriteString(prevLine)
				lastBlank = false
				prevLine = ""
			}
			text.WriteString(line)
		}
	}
	return []byte(text.String()), nil
}

//readRecordValueText if args passed to this function it reads content from that file
//else it asks user to enter path to the file
func readRecordValueFile(reader *bufio.Reader, args ...string) ([]byte, error) {
	var (
		err      error
		filepath string
		content  []byte
	)

	if len(args) > 0 {
		filepath = args[0]

		content, err = os.ReadFile(filepath)
		if err != nil {
			return nil, err
		}
	} else {
		for {
			fmt.Print("Enter path to the file: ")
			filepath, err = reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			filepath = strings.TrimSpace(filepath)

			content, err = os.ReadFile(filepath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("file not found: ", filepath)
					continue
				} else {
					return nil, err
				}
			}
			break
		}
	}

	return models.EncodeFileDataType(&models.FileDataType{
		Filename: path.Base(filepath),
		Contents: content,
	})
}

func readRecordValueLogin(reader *bufio.Reader, args ...string) ([]byte, error) {
	var (
		err      error
		login    string
		password string
	)

	if len(args) < 1 {
		fmt.Print("Enter login: ")
		login, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		login = strings.TrimSpace(login)
	} else {
		login = strings.TrimSpace(args[0])
	}

	if len(args) < 2 {
		fmt.Print("Enter password: ")
		_pass, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return nil, err
		}
		fmt.Print("\n")
		password = string(_pass)
	} else {
		password = args[1]
	}

	return models.EncodeLoginDataType(&models.LoginDataType{
		Login:    login,
		Password: password,
	})
}

func readRecordValueBank(reader *bufio.Reader, args ...string) ([]byte, error) {

	return []byte{}, nil
}

func validateName(name string) error {
	if strings.ContainsRune(name, '.') {
		return errors.New("name cannot contain \".\" symbol")
	}
	return nil
}

func readRecordName(reader *bufio.Reader) (string, error) {
	for {
		fmt.Print("Enter record name: ")
		name, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		name = strings.TrimSpace(name)

		if err = validateName(name); err != nil {
			fmt.Println(err.Error())
			continue
		}

		return name, nil
	}
}

func readRecordType(reader *bufio.Reader) (string, error) {
	for {
		fmt.Println("Record types")
		for i, t := range storage.SupportedTypes {
			fmt.Printf("  [%d] %s\n", i+1, t)
		}
		fmt.Print("Enter record type: ")
		t, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		t = strings.TrimSpace(t)
		if slices.Contains(storage.SupportedTypes, t) {
			return t, nil
		}
		num, err := strconv.Atoi(t)
		if err != nil {
			fmt.Printf("unknown record type (%s)\n", t)
			continue
		}
		if num < 1 || num > len(storage.SupportedTypes) {
			fmt.Printf("enter name of record type or it's number 1-%d\n", len(storage.SupportedTypes))
			continue
		}

		return storage.SupportedTypes[num-1], nil
	}
}
