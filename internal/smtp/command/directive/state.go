package directive

import (
	"fmt"
	"net/textproto"
	"regexp"
	"strings"

	//"github.com/andygrunwald/go-jira"

	"github.com/legionus/jiramail/internal/jiraplus"
)

var (
	re = regexp.MustCompile(`[\s-]+`)
)

func replaceStringTrash(s string) string {
	return strings.ToLower(re.ReplaceAllString(strings.TrimSpace(s), " "))
}

func commandState(client *jiraplus.Client, header textproto.MIMEHeader, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("more arguments required for 'state' command")
	}

	state := replaceStringTrash(strings.Join(args, " "))

	issueID := header.Get("X-Issue-Id")

	transitions, _, err := client.Issue.GetTransitions(issueID)
	if err != nil {
		return fmt.Errorf("unable to get transitions for issue %s: %s", issueID, err)
	}

	var available []string

	for _, transition := range transitions {
		if replaceStringTrash(transition.Name) != state {
			available = append(available, transition.Name)
			continue
		}

		_, err := client.Issue.DoTransition(issueID, transition.ID)
		if err != nil {
			return fmt.Errorf("unable to do transition of issue %s: %s", issueID, err)
		}

		return nil
	}

	return fmt.Errorf("unknown state %q, available: %q", state, available)
}
