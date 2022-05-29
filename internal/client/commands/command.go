package commands

import (
	"context"
	"errors"
)

var ErrNotAuthorized = errors.New("not authorized")

type Command interface {
	GetName() string
	Handle(ctx context.Context, args []string) error
}
