package jiraconv

import (
	"fmt"
	"net/mail"
	"net/textproto"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"

	"github.com/legionus/jirasync/internal/message"
	"github.com/legionus/jirasync/internal/smtp/command"
)

func BoardMessageID(data *jira.Board) string {
	boardID := map[string]string{
		"ID": fmt.Sprintf("%d", data.ID),
	}
	return message.EncodeMessageID("board.jira", boardID)
}

func (c *Converter) Board(data *jira.Board, refs []string) (*mail.Message, error) {
	if data == nil {
		return nil, fmt.Errorf("unable to convert nil to board message")
	}

	headers := make(textproto.MIMEHeader)

	headers.Set("Message-ID", BoardMessageID(data))
	headers.Set("Reply-To", "nobody@jira")
	headers.Set("Date", time.Time(time.Time{}).Format(time.RFC1123Z))
	headers.Set("From", NobodyUser.String())
	headers.Set("Subject", fmt.Sprintf("%s (%d)", data.Name, data.ID))

	if len(refs) > 0 {
		headers.Set("In-Reply-To", refs[len(refs)-1])
		headers.Set("References", strings.Join(refs, " "))
	}

	headers.Set("X-Jira-ID", fmt.Sprintf("%d", data.ID))
	headers.Set("X-Jira-Type", data.Type)

	return &mail.Message{
		Header: mail.Header(headers),
		Body:   strings.NewReader(command.MakeJiraBlock(nil)),
	}, nil
}
