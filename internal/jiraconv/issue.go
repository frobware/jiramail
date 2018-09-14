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

func (c *Converter) processIssue(mType, ID, Key string, fields *jira.IssueFields, refs []string) ([]*mail.Message, error) {
	assignee := UserFromJira(fields.Assignee)
	if assignee != nil {
		c.usercache.Set(fields.Assignee)
	} else {
		assignee = NobodyUser
	}

	creator := UserFromJira(fields.Creator)
	if creator != nil {
		c.usercache.Set(fields.Creator)
	} else {
		creator = NobodyUser
	}

	c.usercache.Set(fields.Reporter)

	date := time.Time(fields.Created)
	updated := time.Time(fields.Updated)

	if !time.Time(time.Time{}).Equal(updated) {
		date = updated
	}

	issueID := map[string]string{
		"ID":  ID,
		"Key": Key,
	}

	headers := make(textproto.MIMEHeader)

	headers.Set("Message-ID", message.EncodeMessageID(mType+".jira", issueID))
	headers.Set("Reply-To", "reply@jira")
	headers.Set("Date", date.Format(time.RFC1123Z))
	headers.Set("From", creator.String())
	headers.Set("To", assignee.String())
	headers.Set("Subject", fmt.Sprintf("[%s] %s", Key, fields.Summary))

	if len(refs) > 0 {
		headers.Set("In-Reply-To", refs[len(refs)-1])
		headers.Set("References", strings.Join(refs, " "))
	}

	if fields.Watches != nil {
		for _, w := range fields.Watches.Watchers {
			watcher, err := c.usercache.Get(w.Name)
			if err != nil {
				return nil, err
			}
			headers.Add("Cc", UserFromJira(watcher).String())
		}
	}

	var issueInfo [][]string

	issueInfo = append(issueInfo, []string{"Type", fields.Type.Name})

	for _, component := range fields.Components {
		issueInfo = append(issueInfo, []string{"Component", component.Name})
	}
	if len(fields.Labels) > 0 {
		issueInfo = append(issueInfo, []string{"Labels", strings.Join(fields.Labels, ", ")})
	}
	if fields.Priority != nil {
		issueInfo = append(issueInfo, []string{"Priority", fields.Priority.Name})
	}
	if fields.Resolution != nil {
		issueInfo = append(issueInfo, []string{"Resolution", fields.Resolution.Name})
	}
	if fields.Status != nil {
		issueInfo = append(issueInfo, []string{"Status", fields.Status.Name})
	}

	for _, field := range c.jiraFields {
		if !field.Custom || !field.Navigable || fields.Unknowns[field.ID] == nil {
			continue
		}
		switch field.Schema.Type {
		case "number", "string":
		default:
			continue
		}
		issueInfo = append(issueInfo, []string{field.Name, fmt.Sprintf("%v", fields.Unknowns[field.ID])})
	}

	ret := []*mail.Message{
		{
			Header: mail.Header(headers),
			Body:   strings.NewReader(command.MakeJiraBlock(issueInfo) + fields.Description),
		},
	}

	var comments []*jira.Comment

	if fields.Comments != nil {
		comments = (*fields.Comments).Comments
	}

	refs = append(refs, headers.Get("Message-ID"))

	for _, comment := range comments {
		var dateStr string

		if len(comment.Updated) > 0 {
			dateStr = comment.Updated
		} else {
			dateStr = comment.Created
		}

		date, err := time.Parse("2006-01-02T15:04:05.999999999-0700", dateStr)
		if err != nil {
			panic(fmt.Sprintf("unable to parse Created field: %s", err))
		}

		commentID := map[string]string{
			"ID": comment.ID,
		}

		cheaders := make(textproto.MIMEHeader)

		cheaders.Set("Message-ID", message.EncodeMessageID("comment.jira", commentID))
		cheaders.Set("Reply-To", "reply@jira")
		cheaders.Set("Date", date.Format(time.RFC1123Z))
		cheaders.Set("From", UserFromJira(&comment.Author).String())
		cheaders.Set("To", assignee.String())
		cheaders.Set("Subject", fmt.Sprintf("%s", headers.Get("Subject")))

		if len(refs) > 0 {
			cheaders.Set("In-Reply-To", refs[len(refs)-1])
			cheaders.Set("References", strings.Join(refs, " "))
		}

		m := &mail.Message{
			Header: mail.Header(cheaders),
			Body:   strings.NewReader(comment.Body),
		}

		ret = append(ret, m)
	}

	for _, subtask := range fields.Subtasks {
		mails, err := c.processIssue("subtask", subtask.ID, subtask.Key, &subtask.Fields, refs)
		if err != nil {
			return nil, err
		}
		for _, m := range mails {
			ret = append(ret, m)
		}
	}

	return ret, nil
}

func (c *Converter) Issue(data *jira.Issue, refs []string) ([]*mail.Message, error) {
	if data == nil || data.Fields == nil {
		return nil, fmt.Errorf("unable to convert nil to issue messages")
	}
	if data.Fields.Type.Subtask {
		return c.processIssue("subtask", data.ID, data.Key, data.Fields, refs)
	}
	return c.processIssue("issue", data.ID, data.Key, data.Fields, refs)
}
