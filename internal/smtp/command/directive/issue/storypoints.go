package issue

import (
	"fmt"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/legionus/jiramail/internal/jiraplus"
	"github.com/legionus/jiramail/internal/smtp/command"
)

func StoryPoints(client *jiraplus.Client, hdr textproto.MIMEHeader, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("wrong number of arguments to set story points")
	}

	points, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("arguments is not number: %s", args[0])
	}

	fields, _, err := client.Field.GetList()

	fieldType := ""
	fieldID := ""
	for _, v := range fields {
		if !v.Custom {
			continue
		}
		if strings.Replace(strings.ToLower(v.Name), " ", "", -1) == "storypoints" {
			if len(fieldID) > 0 {
				return fmt.Errorf("There were more than one field that looked like 'story points'")
			}
			fieldID = v.ID
			fieldType = strings.ToLower(v.Schema.Type)
		}
	}

	if len(fieldID) == 0 {
		return fmt.Errorf("'story points' field not found")
	}

	var data command.JiraMap

	switch fieldType {
	case "number":
		data = command.JiraMap{"update": command.JiraMap{fieldID: points}}
	case "string":
		data = command.JiraMap{"update": command.JiraMap{fieldID: fmt.Sprintf("%d", points)}}
	default:
		return fmt.Errorf("unsupported type of 'story points' field: %s", fieldType)
	}

	issueID := hdr.Get("X-Issue-Id")

	_, err = client.Issue.UpdateIssue(issueID, data)
	if err != nil {
		return fmt.Errorf("unable to update issue %s: %s", issueID, err)
	}

	return nil
}
