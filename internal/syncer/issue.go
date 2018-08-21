package syncer

import (
	"path"

	"github.com/sirupsen/logrus"

	"github.com/andygrunwald/go-jira"
	"github.com/legionus/jirasync/internal/jiraconv"
	"github.com/legionus/jirasync/internal/jiraplus"
	"github.com/legionus/jirasync/internal/maildir"
)

func (s *JiraSyncer) issue(mdir maildir.Dir, issue *jira.Issue, refs []string) error {
	mailList, err := jiraconv.NewConverter(s.remote, s.usercache).Issue(issue, refs)
	if err != nil {
		return err
	}

	for _, msg := range mailList {
		err = s.writeMessage(mdir, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *JiraSyncer) projectissue(issue *jira.Issue) error {
	mdir, err := Maildir(path.Join(s.config.Remote[s.remote].DestDir, "projects", issue.Fields.Project.Key))
	if err != nil {
		return err
	}
	s.projects[issue.Fields.Project.Key] = struct{}{}

	refs := []string{
		jiraconv.RemoteMessageID(s.remote),
		jiraconv.ProjectMessageID(&issue.Fields.Project),
	}

	return s.issue(mdir, issue, refs)
}

func (s *JiraSyncer) issues(mdir maildir.Dir, query string, refs []string) error {
	opts := &jira.SearchOptions{
		StartAt:    0,
		MaxResults: 100,
		Fields:     []string{"*all"},
	}

	count, err := jiraplus.List(
		func(i int) ([]interface{}, error) {
			opts.StartAt = i

			ret, _, err := s.client.Issue.Search(query, opts)
			if err != nil {
				return nil, err
			}

			a := make([]interface{}, len(ret))

			for k := range ret {
				a[k] = &ret[k]
			}

			return a, nil
		},
		func(o interface{}) error {
			issue := o.(*jira.Issue)

			//logrus.Infof("issue %s", issue.Key)

			if err := s.issue(mdir, issue, refs); err != nil {
				return err
			}

			if err := s.projectissue(issue); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	logrus.Infof("issues %d", count)

	return nil
}
