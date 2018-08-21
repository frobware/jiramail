package directive

import (
	"fmt"
	"net/textproto"
	"strings"

	//"github.com/andygrunwald/go-jira"

	"github.com/legionus/jirasync/internal/jiraplus"
	"github.com/legionus/jirasync/internal/smtp/command"
)

func processLabels(client *jiraplus.Client, header textproto.MIMEHeader, cmd string, args []string) error {
	issueID := header.Get("X-Issue-Id")

	var labels []command.JiraMap

	for _, label := range args {
		labels = append(labels, command.JiraMap{cmd: label})
	}

	if len(labels) == 0 {
		return nil
	}

	issue := command.JiraMap{"update": command.JiraMap{"labels": labels}}

	_, err := client.Issue.UpdateIssue(issueID, issue)
	if err != nil {
		return fmt.Errorf("unable to update issue %s: %s", issueID, err)
	}

	return nil
}

func commandLabels(client *jiraplus.Client, hdr textproto.MIMEHeader, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("More arguments required for 'labels' command")
	}

	cmd := strings.ToLower(args[0])

	switch cmd {
	case "add", "remove":
		return processLabels(client, hdr, cmd, args[1:])
	default:
		return fmt.Errorf("command labels got unknown subcommand: %s", args[0])
	}

	return nil
}
