package modify

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/client/storage"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"golang.org/x/exp/slices"
	"golang.org/x/term"
	"os"
	"path"
	"strconv"
	"strings"
)

//readRecordValueText if args passed to this function it returns first element as []byte and array of unused args
//else it asks user to enter text and reads line by line. When two empty line met text considered finished (last two
//empty lines not included to result value)
func readRecordValueText(reader *bufio.Reader, args ...string) ([]byte, []string, error) {
	var text []byte
	if len(args) < 1 {
		lastBlank := false
		prevLine := ""
		fmt.Println("Enter text (leave two blank lines to finish text)")
		var tb = strings.Builder{}
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, args, err
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

	return text, args, nil
}

//readRecordComment if args passed to this function it returns first element as []byte and array of unused args
//else it asks user to enter comment
func readRecordComment(reader *bufio.Reader, args ...string) (string, []string, error) {
	var (
		comment string
		err     error
	)

	if len(args) < 1 {
		fmt.Print("Enter record comment (can be empty): ")
		comment, err = reader.ReadString('\n')
		if err != nil {
			return "", args, err
		}
	} else {
		comment = args[0]
		args = args[1:]
	}

	return strings.TrimSpace(comment), args, nil
}

//readRecordValueFile if args passed to this function it reads content from that file
//else it asks user to enter path to the file
func readRecordValueFile(reader *bufio.Reader, args ...string) ([]byte, []string, error) {
	var (
		err      error
		filepath string
		content  []byte
	)

	if len(args) < 1 {
		for {
			fmt.Print("Enter path to the file: ")
			filepath, err = reader.ReadString('\n')
			if err != nil {
				return nil, args, err
			}
			filepath = strings.TrimSpace(filepath)

			content, err = os.ReadFile(filepath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("file not found: ", filepath)
					continue
				} else {
					return nil, args, err
				}
			}
			break
		}
	} else {
		filepath = args[0]

		content, err = os.ReadFile(filepath)
		if err != nil {
			return nil, args, err
		}
		args = args[1:]
	}

	data, err := models.EncodeFileDataType(&models.FileDataType{
		Filename: path.Base(filepath),
		Contents: content,
	})
	if err != nil {
		return nil, args, err
	}

	return data, args, err
}

//readRecordValueFile if args passed to this function it reads login and password from them
//else it asks user to enter login and password
func readRecordValueLogin(reader *bufio.Reader, args ...string) ([]byte, []string, error) {
	var (
		err      error
		login    string
		password string
	)

	login, args, err = readRecordValueLoginLogin(reader, args...)
	if err != nil {
		return nil, args, err
	}
	password, args, err = readRecordValueLoginPassword(args...)
	if err != nil {
		return nil, args, err
	}

	data, err := models.EncodeLoginDataType(&models.LoginDataType{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return nil, args, err
	}

	return data, args, nil
}

//readRecordValueLoginLogin if args passed to this function it reads login from first element and returns it as sting
//and an array of unused args else it asks user to enter login
func readRecordValueLoginLogin(reader *bufio.Reader, args ...string) (string, []string, error) {
	var (
		err   error
		login string
	)
	if len(args) < 1 {
		fmt.Print("Enter login: ")
		login, err = reader.ReadString('\n')
		if err != nil {
			return "", args, err
		}
		login = strings.TrimSpace(login)
	} else {
		login = strings.TrimSpace(args[0])
		args = args[1:]
	}
	return login, args, nil
}

//readRecordValueLoginPassword asks user to enter password
func readRecordValueLoginPassword(args ...string) (string, []string, error) {
	var password string

	if len(args) < 1 {
		fmt.Print("Enter password: ")
		_pass, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", args, err
		}
		fmt.Print("\n")
		password = string(_pass)
	} else {
		password = args[0]
		args = args[1:]
	}
	return password, args, nil
}

func readNewRecordValueBankCard(reader *bufio.Reader, args ...string) ([]byte, []string, error) {
	var (
		err        error
		cardNumber string
		cardHolder string
		expMonth   string
		expYear    string
		cvv        string
	)

	cardNumber, err = readCardNumber(reader)
	if err != nil {
		return nil, args, err
	}

	cardHolder, err = readCardHolderName(reader)
	if err != nil {
		return nil, args, err
	}

	expMonth, expYear, err = readCardExpiry(reader)
	if err != nil {
		return nil, args, err
	}

	cvv, err = readCardCVV(reader)
	if err != nil {
		return nil, args, err
	}

	data, err := models.EncodeBankCardDataType(&models.BankCardDataType{
		Number:   cardNumber,
		Holder:   cardHolder,
		ExpMonth: expMonth,
		ExpYear:  expYear,
		CVV:      cvv,
	})
	if err != nil {
		return nil, args, err
	}

	return data, args, nil
}

func readCardNumber(reader *bufio.Reader) (string, error) {
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

func readCardHolderName(reader *bufio.Reader) (string, error) {
	fmt.Print("Enter card holder name (can be empty): ")
	cardHolder, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	cardHolder = strings.TrimSpace(cardHolder)
	return cardHolder, nil
}

func readCardExpiry(reader *bufio.Reader) (string, string, error) {
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

func readCardCVV(reader *bufio.Reader) (string, error) {

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

func readRecordName(reader *bufio.Reader, args ...string) (string, []string, error) {
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

func readRecordType(reader *bufio.Reader, args ...string) (string, []string, error) {
	var (
		typeName string
		err      error
	)
	if len(args) < 1 {
		typeName, args, err = readRecordTypeFromConsole(reader, args...)
		if err != nil {
			return "", nil, err
		}
	} else {
		t := strings.TrimSpace(args[0])
		typeName, err = guessRecordType(t)
		if err != nil {
			errText := fmt.Sprintf("%s\nSupported record types:\n", err.Error())
			for i, t := range storage.SupportedTypes {
				errText += fmt.Sprintf("  [%d] %s\n", i+1, t)
			}
			return "", nil, errors.New(errText)
		}
		args = args[1:]
	}
	return typeName, args, nil
}

func readRecordTypeFromConsole(reader *bufio.Reader, args ...string) (string, []string, error) {
	for {
		fmt.Println("Record types")
		for i, t := range storage.SupportedTypes {
			fmt.Printf("  [%d] %s\n", i+1, t)
		}
		fmt.Printf("enter name of record type or it's number 1-%d\n", len(storage.SupportedTypes))
		fmt.Print("Enter record type: ")
		t, err := reader.ReadString('\n')
		if err != nil {
			return "", args, err
		}
		typeName, err := guessRecordType(t)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		return typeName, args, nil
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
