package jiraconv

import (
	"fmt"
	"net/mail"
	"net/textproto"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"

	"github.com/legionus/jiramail/internal/message"
	"github.com/legionus/jiramail/internal/smtp/command"
)

func SprintMessageID(data *jira.Sprint) string {
	sprintID := map[string]string{
		"ID": fmt.Sprintf("%d", data.ID),
	}
	return message.EncodeMessageID("sprint.jira", sprintID)
}

func (c *Converter) Sprint(data *jira.Sprint, refs []string) (*mail.Message, error) {
	if data == nil {
		return nil, fmt.Errorf("unable to convert nil to sprint message")
	}

	date := time.Time{}

	if data.StartDate != nil {
		date = *data.StartDate
	}

	headers := make(textproto.MIMEHeader)

	headers.Set("Message-ID", SprintMessageID(data))
	headers.Set("Reply-To", "nobody@jira")
	headers.Set("Date", date.Format(time.RFC1123Z))
	headers.Set("From", NobodyUser.String())
	headers.Set("Subject", fmt.Sprintf("[%s] %s", strings.ToUpper(data.State), data.Name))

	if len(refs) > 0 {
		headers.Set("In-Reply-To", refs[len(refs)-1])
		headers.Set("References", strings.Join(refs, " "))
	}

	var info [][]string

	if data.StartDate != nil {
		info = append(info, []string{"Date start", data.StartDate.Format(time.RFC1123Z)})
	}
	if data.EndDate != nil {
		info = append(info, []string{"Date end", data.EndDate.Format(time.RFC1123Z)})
	}
	if data.CompleteDate != nil {
		info = append(info, []string{"Date complete", data.CompleteDate.Format(time.RFC1123Z)})
	}

	return &mail.Message{
		Header: mail.Header(headers),
		Body:   strings.NewReader(command.MakeJiraBlock(info)),
	}, nil
}
