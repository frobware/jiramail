package command

import (
	"github.com/legionus/jirasync/internal/config"
)

type Handler interface {
	Handle(cfg *config.Configuration, msg *Mail) error
}

type ErrCommand struct {
	Message string
}

func (e *ErrCommand) Error() string {
	return e.Message
}

type JiraMap map[string]interface{}
