package reply

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/andygrunwald/go-jira"
	//log "github.com/sirupsen/logrus"

	"github.com/legionus/jirasync/internal/client"
	"github.com/legionus/jirasync/internal/config"
	"github.com/legionus/jirasync/internal/smtp/command"
	"github.com/legionus/jirasync/internal/smtp/command/factory"
)

var _ command.Handler = &Reply{}

func init() {
	factory.Register("reply@jira", &Reply{})
}

type Reply struct{}

func (r *Reply) Handle(cfg *config.Configuration, msg *command.Mail) error {
	msgType := msg.Header.Get("X-Type")

	switch msgType {
	case "issue", "comment":
		return addComment(cfg, msg)
	case "project":
		return addIssue(cfg, msg, len(msg.Header.Get("X-Subtask")) > 0)
	}

	return fmt.Errorf("unsupported message type: %s", msgType)
}

func addComment(cfg *config.Configuration, msg *command.Mail) error {
	client, err := client.NewClient(cfg, msg.Header.Get("X-Remote-Name"))
	if err != nil {
		return err
	}

	msgID := msg.Header.Get("X-Issue-Key")

	_, _, err = client.PlusIssue.AddComment(msgID, &jira.Comment{Body: command.GetBody(msg)})
	if err != nil {
		return fmt.Errorf("unable to add comment to %s issue: %s", msgID, err)
	}

	return nil
}

func getIssueType(s string, project *jira.Project) (string, string, error) {
	if len(project.IssueTypes) == 0 {
		return "", "", fmt.Errorf("issue types do not exist")
	}

	res := regexp.MustCompile(`(?i:\[(?:\?|JIRA)\s*TYPE\s+([^\]]+)\])`).FindStringSubmatch(s)
	if len(res) == 0 {
		return project.IssueTypes[0].Name, s, nil
	}

	for _, v := range project.IssueTypes {
		if strings.EqualFold(res[1], v.Name) {
			s = strings.Replace(s, res[0], "", -1)
			return v.Name, s, nil
		}
	}

	return "", "", fmt.Errorf("issue type %s not found", res[1])
}

func addIssue(cfg *config.Configuration, msg *command.Mail, subtask bool) error {
	client, err := client.NewClient(cfg, msg.Header.Get("X-Remote-Name"))
	if err != nil {
		return err
	}

	project, _, err := client.Project.Get(msg.Header.Get("X-Project-Id"))
	if err != nil {
		return fmt.Errorf("unable to get project %s: %s", msg.Header.Get("X-Project-Key"), err)
	}

	issueType, subject, err := getIssueType(msg.Header.Get("Subject"), project)
	if err != nil {
		return err
	}

	issue := jira.Issue{
		Fields: &jira.IssueFields{
			Project: *project,
			Type: jira.IssueType{
				Name:    issueType,
				Subtask: subtask,
			},
			Summary:     subject,
			Description: command.GetBody(msg),
		},
	}

	_, _, err = client.PlusIssue.Create(&issue)
	if err != nil {
		return fmt.Errorf("unable to add issue: %s", err)
	}

	return nil
}
