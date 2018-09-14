package issue

import (
	"fmt"
	"net/textproto"
	"strings"

	"github.com/sirupsen/logrus"
	//"github.com/andygrunwald/go-jira"

	"github.com/legionus/jiramail/internal/jiraplus"
	"github.com/legionus/jiramail/internal/smtp/command"
)

func Priority(client *jiraplus.Client, header textproto.MIMEHeader, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("more arguments required for 'priority' command")
	}

	newPriority := replaceStringTrash(strings.Join(args, " "))

	issueID := header.Get("X-Issue-Id")

	priorities, _, err := client.Priority.GetList()
	if err != nil {
		return fmt.Errorf("unable to get priorities: %s", err)
	}

	var available []string

	for _, priority := range priorities {
		if replaceStringTrash(priority.Name) != newPriority {
			available = append(available, priority.Name)
			continue
		}

		data := command.JiraMap{
			"fields": command.JiraMap{
				"priority": command.JiraMap{"id": priority.ID},
			},
		}

		logrus.Debugf("JIRA REQUEST (ISSUE=%s): %#+v", issueID, data)

		_, err := client.Issue.UpdateIssue(issueID, data)
		if err != nil {
			return fmt.Errorf("unable to set priority %s to issue %s: %s", priority.Name, issueID, err)
		}

		return nil
	}

	return fmt.Errorf("unknown priority %q, available: %q", newPriority, available)
}
