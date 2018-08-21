//
// https://developer.atlassian.com/cloud/jira/software/rest
//
package jiraplus

import (
	"github.com/andygrunwald/go-jira"
)

type Client struct {
	*jira.Client
	PlusIssue *IssueService
	PlusBoard *BoardService
}

func NewClient(c *jira.Client) *Client {
	return &Client{
		Client: c,
		PlusBoard: &BoardService{
			BoardService: c.Board,
			client:       c,
		},
		PlusIssue: &IssueService{
			IssueService: c.Issue,
			client:       c,
		},
	}
}
