package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/client/storage"
	"github.com/putalexey/goph-keeper/internal/common/gproto"
	"go.uber.org/zap"
	"golang.org/x/term"
	"os"
	"strings"
)

type Auth struct {
	logger  *zap.SugaredLogger
	remote  gproto.GKServerClient
	storage storage.Storager
}

func NewAuthCommand(logger *zap.SugaredLogger, remote gproto.GKServerClient, storage storage.Storager) *Auth {
	return &Auth{logger: logger, remote: remote, storage: storage}
}

func (c *Auth) GetName() string {
	return "auth"
}

func (c *Auth) Handle(ctx context.Context, args []string) error {
	var (
		err      error
		login    string
		password []byte
	)

	if len(args) > 1 {
		return errors.New("too many arguments\nusage: gk-client auth [login]")
	}
	reader := bufio.NewReader(os.Stdin)
	if len(args) == 0 {
		fmt.Print("Enter login: ")
		login, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
		login = strings.TrimSpace(login)
	} else {
		login = strings.TrimSpace(args[0])
	}

	for len(password) == 0 {
		fmt.Print("Enter password: ")
		password, err = term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		fmt.Print("\n")
	}

	response, err := c.remote.Authorize(ctx, &gproto.AuthorizeRequest{
		Login:    login,
		Password: string(password),
	})
	if err != nil {
		return err
	}
	c.storage.SetToken(response.AuthToken)
	c.storage.SetLogin(response.User.Login)
	fmt.Println("Successful authorized")
	return nil
}
