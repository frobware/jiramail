package jiraplus

import (
	"fmt"

	"github.com/andygrunwald/go-jira"
)

type IssueService struct {
	*jira.IssueService
	client *jira.Client
}

func (s *IssueService) GetComment(issueID, commentID string) (*jira.Comment, *jira.Response, error) {
	apiEndpoint := fmt.Sprintf("rest/api/2/issue/%s/comment/%s", issueID, commentID)

	req, err := s.client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	v := new(jira.Comment)
	resp, err := s.client.Do(req, v)

	return v, resp, err
}

func (s *IssueService) UpdateComment(issueID string, comment *jira.Comment) (*jira.Comment, *jira.Response, error) {
	reqBody := struct {
		Body string `json:"body"`
	}{
		Body: comment.Body,
	}
	apiEndpoint := fmt.Sprintf("rest/api/2/issue/%s/comment/%s", issueID, comment.ID)
	req, err := s.client.NewRequest("PUT", apiEndpoint, reqBody)
	if err != nil {
		return nil, nil, err
	}

	responseComment := new(jira.Comment)
	resp, err := s.client.Do(req, responseComment)
	if err != nil {
		return nil, resp, err
	}

	return responseComment, resp, nil
}
