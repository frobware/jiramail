package issue

import (
	"fmt"
	"net/textproto"
	"strings"

	"github.com/legionus/jiramail/internal/jiraplus"
)

func Watchers(client *jiraplus.Client, hdr textproto.MIMEHeader, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("more arguments required for command")
	}

	issueID := hdr.Get("X-Issue-Id")

	cmd := strings.ToLower(args[0])

	switch cmd {
	case "add":
		for _, userName := range args {
			_, err := client.Issue.AddWatcher(issueID, userName)
			if err != nil {
				return fmt.Errorf("unable to add watcher to issue %s: %s", issueID, err)
			}
		}
	case "remove":
		for _, userName := range args {
			_, err := client.Issue.RemoveWatcher(issueID, userName)
			if err != nil {
				return fmt.Errorf("unable to add watcher from issue %s: %s", issueID, err)
			}
		}
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}

	return nil
}
