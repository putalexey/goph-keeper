package commands

import "context"

type Command interface {
	GetName() string
	Handle(ctx context.Context, args []string) error
}
