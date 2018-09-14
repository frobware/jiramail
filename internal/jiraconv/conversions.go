package jiraconv

import (
	"github.com/andygrunwald/go-jira"

	"github.com/legionus/jiramail/internal/cache"
)

type Converter struct {
	remote     string
	usercache  *cache.User
	jiraFields []jira.Field
}

func NewConverter(remoteName string, usercache *cache.User) *Converter {
	return &Converter{
		remote:    remoteName,
		usercache: usercache,
	}
}

func (c *Converter) getAssignee(data *jira.IssueFields) *User {
	assignee := UserFromJira(data.Assignee)
	if assignee == nil {
		assignee = NobodyUser
	} else {
		c.usercache.Set(data.Assignee)
	}
	return assignee
}

func (c *Converter) SetJiraFields(data []jira.Field) {
	c.jiraFields = data
}
