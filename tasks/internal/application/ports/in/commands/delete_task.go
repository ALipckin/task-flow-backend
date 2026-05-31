package commands

import (
	"context"
)

type DeleteTaskHandler interface {
	Execute(ctx context.Context, cmd DeleteTaskCommand) (bool, error)
}

type DeleteTaskCommand struct {
	ID uint64
}
