package help

import (
	"context"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/client/commands"
	"github.com/putalexey/goph-keeper/internal/client/storage"
	"github.com/putalexey/goph-keeper/internal/common/gproto"
	"go.uber.org/zap"
)

type Help struct {
	logger   *zap.SugaredLogger
	remote   gproto.GKServerClient
	storage  storage.Storager
	commands []commands.Command
}

func NewHelpCommand(logger *zap.SugaredLogger, remote gproto.GKServerClient, storage storage.Storager, commands []commands.Command) *Help {
	return &Help{logger: logger, remote: remote, storage: storage, commands: commands}
}

func (c *Help) GetName() string {
	return "help"
}

func (c *Help) GetFullDescription() string {
	return `Usage: gk-client help [command]`
}

func (c *Help) GetShortDescription() string {
	return "show current help"
}

func (c *Help) Handle(_ context.Context, args []string) error {
	if len(args) < 1 {
		c.mainHelp()
	} else {
		c.commandHelp(args[0])
	}

	return nil
}

func (c *Help) mainHelp() {
	fmt.Print(`
Usage: gk-client <command> [command arguments...]

A passwords and other data manager
For more info about command use: gk-client help <command>

Commands:
`)
	_cmds := append(c.commands, c)
	for _, command := range _cmds {
		fmt.Printf("    %s - %s\n", command.GetName(), command.GetShortDescription())
	}
}

func (c *Help) commandHelp(commandName string) {
	_cmds := append(c.commands, c)
	for _, command := range _cmds {
		//fmt.Printf("%s - %s\n", command.GetName(), command.GetFullDescription())
		if command.GetName() == commandName {
			fmt.Println(command.GetFullDescription())
			return
		}
	}
	fmt.Printf("Command \"%s\" not found\n", commandName)
}
