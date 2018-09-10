package jiraconv

import (
	"fmt"
	"net/mail"
	"net/textproto"
	"sort"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"

	"github.com/legionus/jiramail/internal/message"
	"github.com/legionus/jiramail/internal/smtp/command"
)

func ProjectMessageID(data *jira.Project) string {
	projectID := map[string]string{
		"ID":  data.ID,
		"Key": data.Key,
	}
	return message.EncodeMessageID("project.jira", projectID)
}

func (c *Converter) Project(data *jira.Project, refs []string) (*mail.Message, error) {
	if data == nil {
		return nil, fmt.Errorf("unable to convert nil to project message")
	}

	lead, err := c.usercache.Get(data.Lead.Name)
	if err != nil {
		return nil, err
	}

	headers := make(textproto.MIMEHeader)

	headers.Set("Message-ID", ProjectMessageID(data))
	headers.Set("Reply-To", "reply@jira")
	headers.Set("Date", time.Time(time.Time{}).Format(time.RFC1123Z))
	headers.Set("From", UserFromJira(lead).String())
	headers.Set("Subject", fmt.Sprintf("[%s] %s", data.Key, data.Name))

	if len(refs) > 0 {
		headers.Set("In-Reply-To", refs[len(refs)-1])
		headers.Set("References", strings.Join(refs, " "))
	}

	var info [][]string

	if len(data.ProjectCategory.Name) > 0 {
		info = append(info, []string{"Category", data.ProjectCategory.Name})
	}
	if len(data.Components) > 0 {
		s := make([]string, len(data.Components))
		for i, component := range data.Components {
			s[i] = `"` + component.Name + `"`
		}
		sort.Strings(s)
		info = append(info, []string{"Components", strings.Join(s, ", ")})
	}
	if len(data.IssueTypes) > 0 {
		s := make([]string, len(data.IssueTypes))
		for i, issuetype := range data.IssueTypes {
			s[i] = `"` + issuetype.Name + `"`
		}
		sort.Strings(s)
		info = append(info, []string{"Issue types", strings.Join(s, ", ")})
	}
	if len(data.Email) > 0 {
		info = append(info, []string{"Email", data.Email})
	}
	if len(data.URL) > 0 {
		info = append(info, []string{"URL", data.URL})
	}

	return &mail.Message{
		Header: mail.Header(headers),
		Body:   strings.NewReader(command.MakeJiraBlock(info) + data.Description),
	}, nil
}
