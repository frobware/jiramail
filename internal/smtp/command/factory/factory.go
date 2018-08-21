package factory

import (
	"fmt"

	"github.com/legionus/jirasync/internal/smtp/command"
)

var (
	InvalidMailHandlerError = fmt.Errorf("Command handler not registered")

	mailHandler = make(map[string]command.Handler)
)

func Register(name string, handler command.Handler) {
	if handler == nil {
		panic("Must not provide nil command handler")
	}
	_, registered := mailHandler[name]
	if registered {
		panic(fmt.Sprintf("Command handler named %s already registered", name))
	}

	mailHandler[name] = handler
}

func Get(name string) (command.Handler, error) {
	handler, ok := mailHandler[name]
	if !ok {
		return nil, InvalidMailHandlerError
	}
	return handler, nil
}
