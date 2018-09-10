package replace

import (
	"fmt"
	"strings"

	//log "github.com/sirupsen/logrus"
	"github.com/andygrunwald/go-jira"

	"github.com/legionus/jiramail/internal/client"
	"github.com/legionus/jiramail/internal/config"
	"github.com/legionus/jiramail/internal/jiraplus"
	"github.com/legionus/jiramail/internal/smtp/command"
	"github.com/legionus/jiramail/internal/smtp/command/factory"
)

var _ command.Handler = &ReplaceBody{}

func init() {
	handler := &ReplaceBody{}
	factory.Register("edit@jira", handler)
	factory.Register("change@jira", handler)
	factory.Register("replace@jira", handler)
}

type ReplaceBody struct{}

func replaceIssue(client *jiraplus.Client, msg *command.Mail) error {
	issueID := msg.Header.Get("X-Issue-Id")
	issueKey := msg.Header.Get("X-Issue-Key")

	subject := strings.TrimSpace(msg.Header.Get("Subject"))

	subject = strings.TrimSpace(strings.TrimPrefix(subject, "Re:")) // RFC5322
	subject = strings.TrimSpace(strings.TrimPrefix(subject, "re:"))
	subject = strings.TrimSpace(strings.TrimPrefix(subject, "["+issueKey+"]"))

	issue := command.JiraMap{
		"update": command.JiraMap{
			"summary":     []command.JiraMap{{"set": subject}},
			"description": []command.JiraMap{{"set": command.GetBody(msg)}},
		},
	}

	_, err := client.Issue.UpdateIssue(issueID, issue)
	if err != nil {
		return fmt.Errorf("unable to update issue %s: %s", issueID, err)
	}

	return nil
}

func replaceComment(client *jiraplus.Client, msg *command.Mail) error {
	commentID := msg.Header.Get("X-Comment-Id")
	issueID := msg.Header.Get("X-Issue-Key")

	comment := &jira.Comment{
		ID:   commentID,
		Body: command.GetBody(msg),
	}

	_, _, err := client.PlusIssue.UpdateComment(issueID, comment)
	if err != nil {
		return fmt.Errorf("unable to update comment %s in the issue %s: %s", commentID, issueID, err)
	}

	return nil
}

func (e *ReplaceBody) Handle(cfg *config.Configuration, msg *command.Mail) error {
	msgType := msg.Header.Get("X-Type")
	msgRemote := msg.Header.Get("X-Remote-Name")

	client, err := client.NewClient(cfg, msgRemote)
	if err != nil {
		return err
	}

	switch msgType {
	case "issue":
		return replaceIssue(client, msg)
	case "comment":
		return replaceComment(client, msg)
	default:
		return fmt.Errorf("unsupported message type: %s", msgType)
	}
	return nil
}
