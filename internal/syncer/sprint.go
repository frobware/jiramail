package syncer

import (
	"fmt"
	"path"

	"github.com/andygrunwald/go-jira"
	"github.com/sirupsen/logrus"

	"github.com/legionus/jirasync/internal/jiraconv"
	"github.com/legionus/jirasync/internal/jiraplus"
	"github.com/legionus/jirasync/internal/maildir"
)

func (s *JiraSyncer) sprint(parent maildir.Dir, board *jira.Board, sprint *jira.Sprint, refs []string) error {
	// TODO(legion) Fix this mess

	logrus.Infof("sprint %q", sprint.Name)

	msg, err := jiraconv.NewConverter(s.remote, s.usercache).Sprint(sprint, append(refs, jiraconv.BoardMessageID(board)))
	if err != nil {
		return err
	}

	// write sprint to the board maildir
	err = s.writeMessage(parent, msg)
	if err != nil {
		return err
	}

	mdir, err := Maildir(path.Join(string(parent), fmt.Sprintf("%s (%d)", ReplaceStringTrash(sprint.Name), sprint.ID)))
	if err != nil {
		return err
	}

	// write sprint to the sprint maildir
	err = s.writeMessage(mdir, msg)
	if err != nil {
		return err
	}

	// write board to the sprint maildir
	msg, err = jiraconv.NewConverter(s.remote, s.usercache).Board(board, refs)
	if err != nil {
		return err
	}

	err = s.writeMessage(mdir, msg)
	if err != nil {
		return err
	}

	refs = append(refs, jiraconv.BoardMessageID(board), jiraconv.SprintMessageID(sprint))

	opts := &jiraplus.BoardIssuesSearchOptions{}
	opts.MaxResults = 100

	count, err := jiraplus.List(
		func(i int) ([]interface{}, error) {
			opts.StartAt = i

			ret, _, err := s.client.PlusBoard.GetIssuesForSprint(sprint.OriginBoardID, sprint.ID, opts)
			if err != nil {
				return nil, fmt.Errorf("unable to get issues for sprint '%s': %s", sprint.Name, err)
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

	logrus.Infof("sprint issues %d", count)

	// Garbage collection
	err = s.CleanDir(mdir)
	if err != nil {
		return err
	}

	return nil
}

func (s *JiraSyncer) sprints(parent maildir.Dir, board *jira.Board, refs []string) error {
	opts := &jira.GetAllSprintsOptions{}
	opts.MaxResults = 100

	count, err := jiraplus.List(
		func(i int) ([]interface{}, error) {
			opts.StartAt = i

			ret, _, err := s.client.Board.GetAllSprintsWithOptions(board.ID, opts)
			if err != nil {
				return nil, err
			}

			a := make([]interface{}, len(ret.Values))

			for k := range ret.Values {
				a[k] = &ret.Values[k]
			}

			return a, nil
		},
		func(o interface{}) error {
			sprint := o.(*jira.Sprint)

			err := s.sprint(parent, board, sprint, refs)
			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	logrus.Infof("board sprints %d", count)

	return nil
}
