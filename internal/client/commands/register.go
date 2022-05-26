package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/client/storage"
	proto "github.com/putalexey/goph-keeper/internal/common/gproto"
	"go.uber.org/zap"
	"golang.org/x/term"
	"os"
	"strings"
)

type RegisterCommand struct {
	logger  *zap.SugaredLogger
	remote  proto.GKServerClient
	storage storage.Storager
}

func NewRegisterCommand(logger *zap.SugaredLogger, remote proto.GKServerClient, storage storage.Storager) *RegisterCommand {
	return &RegisterCommand{logger: logger, remote: remote, storage: storage}
}

func (c *RegisterCommand) GetName() string {
	return "register"
}

func (c *RegisterCommand) Handle(ctx context.Context, args []string) error {
	var (
		err      error
		login    string
		password []byte
	)

	if len(args) > 1 {
		return errors.New("too many arguments\nusage: gk-client register [login]")
	}
	reader := bufio.NewReader(os.Stdin)
	if len(args) == 0 {
		fmt.Print("Enter login: ")
		login, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
		login = strings.TrimSpace(login)
	}
	fmt.Print("Enter password: ")
	password, err = term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	fmt.Print("\n")

	if len(password) < 4 {
		fmt.Println("password is too short, minimum length is 4")
		return nil
	}

	response, err := c.remote.Register(ctx, &proto.RegisterRequest{
		Login:    login,
		Password: string(password),
	})
	if err != nil {
		return err
	}
	c.storage.SetToken(response.AuthToken)
	c.storage.SetLogin(response.User.Login)
	fmt.Println("Successful registered")
	return nil
}
