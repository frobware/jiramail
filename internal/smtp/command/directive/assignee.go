package directive

import (
	"fmt"
	"net/textproto"
	"strings"

	"github.com/andygrunwald/go-jira"

	"github.com/legionus/jiramail/internal/jiraplus"
)

func commandAssignee(client *jiraplus.Client, header textproto.MIMEHeader, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("more arguments required for 'assignee' command")
	}

	var (
		user *jira.User
		err  error
	)

	if len(args) > 1 {
		if strings.EqualFold(args[0], "to") {
			args = args[1:]
		}
		if len(args) > 1 {
			return fmt.Errorf("too many arguments for 'assignee' command")
		}
	}

	if !strings.EqualFold(args[0], "me") {
		users, _, err := client.User.Find(args[0])
		if err != nil {
			return fmt.Errorf("unable to find user: %s", err)
		}

		if len(users) != 1 {
			return fmt.Errorf("too many users found for pattern: %s", args[0])
		}

		user = &users[0]
	} else {
		user, _, err = client.User.GetSelf()
		if err != nil {
			return fmt.Errorf("unable to get current user: %s", err)
		}
	}

	issueID := header.Get("X-Issue-Id")

	_, err = client.Issue.UpdateAssignee(issueID, user)
	if err != nil {
		return fmt.Errorf("unable to assignee issue %s to %s: %s", issueID, user.Name, err)
	}

	return nil
}
