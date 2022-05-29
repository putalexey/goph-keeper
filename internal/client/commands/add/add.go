package add

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/client/storage"
	"github.com/putalexey/goph-keeper/internal/common/gproto"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"golang.org/x/term"
	"os"
	"path"
	"strconv"
	"strings"
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

func (c *Add) GetHelp() string {
	return `add new record syntax:
gk-client add
gk-client add text [record_name] [text] [comment]
gk-client add file [record_name] [filepath] [comment]
gk-client add login [record_name] [login] [password] [comment]
gk-client add card [record_name]`
}

func (c *Add) Handle(ctx context.Context, args []string) error {
	var (
		err      error
		name     string
		typeName string
		data     []byte
		comment  string
	)

	reader := bufio.NewReader(os.Stdin)

	if len(args) < 1 {
		typeName, err = readRecordType(reader)
		if err != nil {
			return err
		}
	} else {
		t := strings.TrimSpace(args[0])
		typeName, err = guessRecordType(t)
		if err != nil {
			errText := fmt.Sprintf("%s\nSupported record types:\n", err.Error())
			for i, t := range storage.SupportedTypes {
				errText += fmt.Sprintf("  [%d] %s\n", i+1, t)
			}
			return errors.New(errText)
		}
		args = args[1:]
	}

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

	if len(args) < 1 {
		data, comment, err = readRecordValue(typeName, reader)
	} else {
		data, comment, err = readRecordValue(typeName, reader, args...)
	}
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

	//c.storage.SetToken(response.AuthToken)
	//c.storage.SetLogin(response.User.Login)
	fmt.Println("Successful added")

	return nil
}

func readRecordValue(typeName string, reader *bufio.Reader, args ...string) ([]byte, string, error) {
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
	return []byte{}, "", nil
}

//readRecordValueText if args passed to this function it returns first element as []byte
//else it asks user to enter text and reads line by line. When two empty line met text considered finished (last two
//empty lines not included to result value)
func readRecordValueText(reader *bufio.Reader, args ...string) ([]byte, string, error) {
	var (
		err     error
		comment string
		text    []byte
	)
	if len(args) < 1 {
		lastBlank := false
		prevLine := ""
		fmt.Println("Enter text (leave two blank lines to finish text)")
		var tb = strings.Builder{}
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, "", err
			}
			if len(strings.TrimSpace(line)) == 0 {
				if lastBlank {
					break
				}
				lastBlank = true
				prevLine = line
			} else {
				if lastBlank {
					tb.WriteString(prevLine)
					lastBlank = false
					prevLine = ""
				}
				tb.WriteString(line)
			}
		}
		text = []byte(tb.String())
	} else {
		text = []byte(args[0])
		args = args[1:]
	}

	if len(args) < 1 {
		fmt.Print("Enter record comment (can be empty): ")
		comment, err = reader.ReadString('\n')
		if err != nil {
			return nil, "", err
		}
	} else {
		comment = args[0]
	}

	return text, strings.TrimSpace(comment), nil
}

//readRecordValueText if args passed to this function it reads content from that file
//else it asks user to enter path to the file
func readRecordValueFile(reader *bufio.Reader, args ...string) ([]byte, string, error) {
	var (
		err      error
		comment  string
		filepath string
		content  []byte
	)

	if len(args) < 1 {
		for {
			fmt.Print("Enter path to the file: ")
			filepath, err = reader.ReadString('\n')
			if err != nil {
				return nil, "", err
			}
			filepath = strings.TrimSpace(filepath)

			content, err = os.ReadFile(filepath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("file not found: ", filepath)
					continue
				} else {
					return nil, "", err
				}
			}
			break
		}
	} else {
		filepath = args[0]

		content, err = os.ReadFile(filepath)
		if err != nil {
			return nil, "", err
		}
		args = args[1:]
	}

	data, err := models.EncodeFileDataType(&models.FileDataType{
		Filename: path.Base(filepath),
		Contents: content,
	})
	if err != nil {
		return nil, "", err
	}

	if len(args) < 1 {
		fmt.Print("Enter record comment (can be empty): ")
		comment, err = reader.ReadString('\n')
		if err != nil {
			return nil, "", err
		}
	} else {
		comment = args[0]
	}

	return data, comment, err
}

func readRecordValueLogin(reader *bufio.Reader, args ...string) ([]byte, string, error) {
	var (
		err      error
		comment  string
		login    string
		password string
	)

	if len(args) < 1 {
		fmt.Print("Enter login: ")
		login, err = reader.ReadString('\n')
		if err != nil {
			return nil, "", err
		}
		login = strings.TrimSpace(login)
	} else {
		login = strings.TrimSpace(args[0])
		args = args[1:]
	}

	if len(args) < 1 {
		fmt.Print("Enter password: ")
		_pass, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return nil, "", err
		}
		fmt.Print("\n")
		password = string(_pass)
	} else {
		password = args[0]
		args = args[1:]
	}

	data, err := models.EncodeLoginDataType(&models.LoginDataType{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return nil, "", err
	}

	if len(args) < 1 {
		fmt.Print("Enter record comment (can be empty): ")
		comment, err = reader.ReadString('\n')
		if err != nil {
			return nil, "", err
		}
	} else {
		comment = args[0]
	}

	return data, comment, nil
}

func readNewRecordValueBankCard(reader *bufio.Reader, args ...string) ([]byte, string, error) {
	var (
		err        error
		comment    string
		cardNumber string
		cardHolder string
		expMonth   string
		expYear    string
		cvv        string
	)

	cardNumber, err = readNewCardNumber(reader)
	if err != nil {
		return nil, "", err
	}

	cardHolder, err = readNewCardHolderName(reader)
	if err != nil {
		return nil, "", err
	}

	expMonth, expYear, err = readNewCardExpiry(reader)
	if err != nil {
		return nil, "", err
	}

	cvv, err = readNewCardCVV(reader)
	if err != nil {
		return nil, "", err
	}

	data, err := models.EncodeBankCardDataType(&models.BankCardDataType{
		Number:   cardNumber,
		Holder:   cardHolder,
		ExpMonth: expMonth,
		ExpYear:  expYear,
		CVV:      cvv,
	})
	if err != nil {
		return nil, "", err
	}

	if len(args) < 1 {
		fmt.Print("Enter record comment (can be empty): ")
		comment, err = reader.ReadString('\n')
		if err != nil {
			return nil, "", err
		}
	} else {
		comment = args[0]
	}

	return data, comment, nil
}

func readNewCardNumber(reader *bufio.Reader) (string, error) {
	for {
		fmt.Print("Enter card number: ")
		cardNumber, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		cardNumber = strings.TrimSpace(cardNumber)
		if len(cardNumber) >= 15 {
			return cardNumber, nil
		}
		fmt.Println("card number is too short")
	}
}

func readNewCardHolderName(reader *bufio.Reader) (string, error) {
	fmt.Print("Enter card holder name (can be empty): ")
	cardHolder, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	cardHolder = strings.TrimSpace(cardHolder)
	return cardHolder, nil
}

func readNewCardExpiry(reader *bufio.Reader) (string, string, error) {
	var (
		expMonth string
		expYear  string
	)

	for {
		fmt.Print("Enter card expiry month (1-12): ")
		month, err := reader.ReadString('\n')
		if err != nil {
			return "", "", err
		}
		month = strings.TrimSpace(month)
		nMonth, err := strconv.Atoi(month)

		if err == nil && nMonth > 0 && nMonth <= 12 {
			expMonth = month
			break
		}
		fmt.Println("enter number between 1 and 12")
	}

	for {
		fmt.Print("Enter card expiry year: ")
		year, err := reader.ReadString('\n')
		if err != nil {
			return "", "", err
		}
		year = strings.TrimSpace(year)
		if len(year) == 2 {
			year = "20" + year
		}
		_, err = strconv.Atoi(year)

		if err == nil {
			expYear = year
			break
		}
		fmt.Println(err)
	}

	return expMonth, expYear, nil
}

func readNewCardCVV(reader *bufio.Reader) (string, error) {

	fmt.Print("Enter card cvv (leave blank if don't want to store it): ")
	cvv, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	cvv = strings.TrimSpace(cvv)
	return cvv, nil
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
		fmt.Printf("enter name of record type or it's number 1-%d\n", len(storage.SupportedTypes))
		fmt.Print("Enter record type: ")
		t, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		typeName, err := guessRecordType(t)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		return typeName, nil
	}
}

func guessRecordType(t string) (string, error) {
	t = strings.TrimSpace(t)
	if slices.Contains(storage.SupportedTypes, t) {
		return t, nil
	}
	num, err := strconv.Atoi(t)
	if err != nil {
		fmt.Printf("unknown record type (%s)", t)
		return "", errors.New(fmt.Sprintf("unknown record type (%s)", t))
	}
	if num < 1 || num > len(storage.SupportedTypes) {
		return "", errors.New(fmt.Sprintf("unknown record type (%s)", t))
	}
	return storage.SupportedTypes[num-1], nil
}
