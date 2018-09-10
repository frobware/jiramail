package directive

import (
	"fmt"
	"net/textproto"
	"strings"

	"github.com/kballard/go-shellquote"
	//log "github.com/sirupsen/logrus"

	"github.com/legionus/jirasync/internal/client"
	"github.com/legionus/jirasync/internal/config"
	"github.com/legionus/jirasync/internal/jiraplus"
	"github.com/legionus/jirasync/internal/smtp/command"
	"github.com/legionus/jirasync/internal/smtp/command/factory"
)

var (
	ErrEndOfCommands = fmt.Errorf("End of commands")
)

func init() {
	handler := &Directive{}
	factory.Register("bot@jira", handler)
	factory.Register("command@jira", handler)
}

func parseCommand(client *jiraplus.Client, msgType string, s string, hdr textproto.MIMEHeader) error {
	args, err := shellquote.Split(s)
	if err != nil {
		return fmt.Errorf("unable to split string: %s", err)
	}

	if len(args) == 0 {
		return nil
	}

	switch strings.ToLower(strings.TrimSpace(args[0])) {
	case "labels":
		if msgType != "issue" {
			return fmt.Errorf("unable to use %q command for this message", args[0])
		}
		return commandLabels(client, hdr, args[1:])
	case "assignee":
		if msgType != "issue" {
			return fmt.Errorf("unable to use %q command for this message", args[0])
		}
		return commandAssignee(client, hdr, args[1:])
	case "state":
		if msgType != "issue" {
			return fmt.Errorf("unable to use %q command for this message", args[0])
		}
		return commandState(client, hdr, args[1:])
	case "end", "--":
		return ErrEndOfCommands
	case "re:", ">", "#", "":
		// ignore
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}

	return nil
}

var _ command.Handler = &Directive{}

type Directive struct{}

func (d *Directive) Handle(cfg *config.Configuration, msg *command.Mail) error {
	msgType := msg.Header.Get("X-Type")
	msgRemote := msg.Header.Get("X-Remote-Name")

	client, err := client.NewClient(cfg, msgRemote)
	if err != nil {
		return err
	}

	err = parseCommand(client, msgType, msg.Header.Get("Subject"), msg.Header)
	if err != nil {
		if err == ErrEndOfCommands {
			return nil
		}
		return err
	}

	for _, line := range msg.Body {
		if err = parseCommand(client, msgType, line, msg.Header); err != nil {
			if err == ErrEndOfCommands {
				return nil
			}
			return err
		}
	}

	return nil
}
