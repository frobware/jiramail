package jiraplus

import (
	"fmt"

	"github.com/andygrunwald/go-jira"
)

type IssueLinkService struct {
	client *jira.Client
}

func (s *IssueLinkService) GetTypes() ([]jira.IssueLinkType, *jira.Response, error) {
	apiEndpoint := fmt.Sprintf("rest/api/2/issueLinkType")

	req, err := s.client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	v := &struct {
		IssueLinkTypes []jira.IssueLinkType `json:"issueLinkTypes" structs:"issueLinkTypes"`
	}{}

	resp, err := s.client.Do(req, v)
	if err != nil {
		err = jira.NewJiraError(resp, err)
	}

	return v.IssueLinkTypes, resp, err
}

func (s *IssueLinkService) AddLink(link *jira.IssueLink) (*jira.Response, error) {
	apiEndpoint := fmt.Sprintf("rest/api/2/issueLink")
	req, err := s.client.NewRequest("POST", apiEndpoint, link)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		err = jira.NewJiraError(resp, err)
	}

	return resp, err
}

func (s *IssueLinkService) DeleteLink(link *jira.IssueLink) (*jira.Response, error) {
	apiEndpoint := fmt.Sprintf("rest/api/2/issueLink/%s", link.ID)
	req, err := s.client.NewRequest("DELETE", apiEndpoint, link)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		err = jira.NewJiraError(resp, err)
	}

	return resp, err
}
