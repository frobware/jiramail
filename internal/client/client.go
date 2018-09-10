package client

import (
	"fmt"
	"net/http"

	"github.com/andygrunwald/go-jira"

	"github.com/legionus/jirasync/internal/config"
	"github.com/legionus/jirasync/internal/jiraplus"
)

func NewClient(c *config.Configuration, remoteName string) (*jiraplus.Client, error) {
	r, ok := c.Remote[remoteName]
	if !ok {
		return nil, fmt.Errorf("remote is not defined in the configuration")
	}

	var httpClient *http.Client

	if len(r.Username) > 0 {
		trans := jira.BasicAuthTransport{
			Username: r.Username,
			Password: r.Password,
		}
		httpClient = trans.Client()
	}

	jiraClient, err := jira.NewClient(httpClient, r.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to create client: %s", err)
	}

	if err == nil {
		return jiraplus.NewClient(jiraClient), nil
	}

	return nil, err
}
