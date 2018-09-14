package sprint

import (
	"fmt"
	"net/textproto"
	"strconv"

	"github.com/legionus/jiramail/internal/jiraplus"
)

func AddIssues(client *jiraplus.Client, hdr textproto.MIMEHeader, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("more arguments required for command")
	}

	sprintID, err := strconv.Atoi(hdr.Get("X-Sprint-Id"))
	if err != nil {
		return err
	}

	_, err = client.Sprint.MoveIssuesToSprint(sprintID, args)
	if err != nil {
		return fmt.Errorf("unable to move issues to sprint: %s", err)
	}

	return nil
}
