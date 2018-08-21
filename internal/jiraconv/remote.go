package jiraconv

import (
	"github.com/legionus/jirasync/internal/message"
)

func RemoteMessageID(name string) string {
	remoteID := map[string]string{
		"Name": name,
	}
	return message.EncodeMessageID("remote.jira", remoteID)
}
