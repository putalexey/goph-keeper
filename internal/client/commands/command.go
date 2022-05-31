package commands

import (
	"context"
	"errors"
)

var ErrNotAuthorized = errors.New("not authorized")

type Command interface {
	GetName() string
	GetFullDescription() string
	GetShortDescription() string
	Handle(ctx context.Context, args []string) error
}
