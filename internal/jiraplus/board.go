package jiraplus

import (
	"fmt"

	"github.com/andygrunwald/go-jira"
)

type BoardIssuesList struct {
	MaxResults int           `json:"maxResults" structs:"maxResults"`
	StartAt    int           `json:"startAt" structs:"startAt"`
	Total      int           `json:"total" structs:"total"`
	Issues     []*jira.Issue `json:"issues" structs:"issues"`
}

type BoardService struct {
	*jira.BoardService
	client *jira.Client
}

type BoardIssuesSearchOptions struct {
	MaxResults    int      `json:"maxResults" structs:"maxResults"`
	StartAt       int      `json:"startAt" structs:"startAt"`
	Jql           string   `json:"jql" structs:"jql"`
	ValidateQuery bool     `json:"validateQuery" structs:"validateQuery"`
	Fields        []string `json:"fields" structs:"fields"`
	Expand        string   `json:"expand" structs:"expand"`
}

func (s *BoardService) getEndpointIssues(apiEndpoint string, opt *BoardIssuesSearchOptions) (*BoardIssuesList, *jira.Response, error) {
	url, err := addOptions(apiEndpoint, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	v := new(BoardIssuesList)
	resp, err := s.client.Do(req, v)
	if err != nil {
		jerr := jira.NewJiraError(resp, err)
		return nil, resp, jerr
	}

	return v, resp, err
}

func (s *BoardService) GetIssues(boardID int, opt *BoardIssuesSearchOptions) (*BoardIssuesList, *jira.Response, error) {
	return s.getEndpointIssues(fmt.Sprintf("rest/agile/1.0/board/%d/issue", boardID), opt)
}

func (s *BoardService) GetIssuesBacklog(boardID int, opt *BoardIssuesSearchOptions) (*BoardIssuesList, *jira.Response, error) {
	return s.getEndpointIssues(fmt.Sprintf("rest/agile/1.0/board/%d/backlog", boardID), opt)
}

func (s *BoardService) GetIssuesWithoutEpic(boardID int, opt *BoardIssuesSearchOptions) (*BoardIssuesList, *jira.Response, error) {
	return s.getEndpointIssues(fmt.Sprintf("rest/agile/1.0/board/%d/epic/none/issue", boardID), opt)
}

func (s *BoardService) GetIssuesForEpic(boardID, epicID int, opt *BoardIssuesSearchOptions) (*BoardIssuesList, *jira.Response, error) {
	return s.getEndpointIssues(fmt.Sprintf("rest/agile/1.0/board/%d/epic/%d/issue", boardID, epicID), opt)
}

func (s *BoardService) GetIssuesForSprint(boardID, sprintID int, opt *BoardIssuesSearchOptions) (*BoardIssuesList, *jira.Response, error) {
	return s.getEndpointIssues(fmt.Sprintf("rest/agile/1.0/board/%d/sprint/%d/issue", boardID, sprintID), opt)
}

type BoardProjectsList struct {
	MaxResults int             `json:"maxResults" structs:"maxResults"`
	StartAt    int             `json:"startAt" structs:"startAt"`
	Total      int             `json:"total" structs:"total"`
	IsLast     bool            `json:"isLast" structs:"isLast"`
	Values     []*jira.Project `json:"values" structs:"values"`
}

func (s *BoardService) GetProjects(boardID int, opt *jira.SearchOptions) (*BoardProjectsList, *jira.Response, error) {
	apiEndpoint := fmt.Sprintf("rest/agile/1.0/board/%d/project", boardID)

	url, err := addOptions(apiEndpoint, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	v := new(BoardProjectsList)
	resp, err := s.client.Do(req, v)
	if err != nil {
		jerr := jira.NewJiraError(resp, err)
		return nil, resp, jerr
	}

	return v, resp, err
}

type BoardEpicsList struct {
	MaxResults int          `json:"maxResults" structs:"maxResults"`
	StartAt    int          `json:"startAt" structs:"startAt"`
	Total      int          `json:"total" structs:"total"`
	IsLast     bool         `json:"isLast" structs:"isLast"`
	Values     []*jira.Epic `json:"values" structs:"values"`
}

func (s *BoardService) GetEpics(boardID int, opt *jira.SearchOptions) (*BoardEpicsList, *jira.Response, error) {
	apiEndpoint := fmt.Sprintf("rest/agile/1.0/board/%d/epic", boardID)

	url, err := addOptions(apiEndpoint, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	v := new(BoardEpicsList)
	resp, err := s.client.Do(req, v)
	if err != nil {
		jerr := jira.NewJiraError(resp, err)
		return nil, resp, jerr
	}

	return v, resp, err
}
