package syncer

import (
	"path"

	"github.com/andygrunwald/go-jira"
	"github.com/sirupsen/logrus"

	"github.com/legionus/jirasync/internal/jiraconv"
	"github.com/legionus/jirasync/internal/jiraplus"
)

func (s *JiraSyncer) backlog(parent string, board *jira.Board, refs []string) error {
	mdir, err := Maildir(path.Join(parent, "Backlog"))
	if err != nil {
		return err
	}

	msg, err := jiraconv.NewConverter(s.remote, s.usercache).Board(board, refs)
	if err != nil {
		return err
	}

	err = s.writeMessage(mdir, msg)
	if err != nil {
		return err
	}

	refs = append(refs, msg.Header.Get("Message-ID"))

	opts := &jiraplus.BoardIssuesSearchOptions{}
	opts.MaxResults = 100

	count, err := jiraplus.List(
		func(i int) ([]interface{}, error) {
			opts.StartAt = i

			ret, _, err := s.client.PlusBoard.GetIssuesBacklog(board.ID, opts)
			if err != nil {
				return nil, err
			}

			if ret.Total <= i {
				return nil, nil
			}

			a := make([]interface{}, len(ret.Issues))

			for k := range ret.Issues {
				a[k] = ret.Issues[k]
			}

			return a, nil
		},
		func(o interface{}) error {
			issue := o.(*jira.Issue)

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

	logrus.Infof("board backlog %d", count)

	// Garbage collection
	err = s.CleanDir(mdir)
	if err != nil {
		return err
	}

	return nil
}
