package syncer

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/andygrunwald/go-jira"
	"github.com/sirupsen/logrus"

	"github.com/legionus/jirasync/internal/jiraconv"
	"github.com/legionus/jirasync/internal/jiraplus"
)

func (s *JiraSyncer) Boards() error {
	var (
		re *regexp.Regexp
	)

	if len(s.config.Remote[s.remote].ProjectMatch) > 0 {
		re = regexp.MustCompile(s.config.Remote[s.remote].BoardMatch)
	}

	refs := []string{jiraconv.RemoteMessageID(s.remote)}

	opts := &jira.BoardListOptions{}
	opts.MaxResults = 100

	handled := 0
	count, err := jiraplus.List(
		func(i int) ([]interface{}, error) {
			opts.StartAt = i

			ret, _, err := s.client.Board.GetAllBoards(opts)
			if err != nil {
				return nil, fmt.Errorf("unable to get boards: %s", err)
			}

			a := make([]interface{}, len(ret.Values))

			for k := range ret.Values {
				a[k] = &ret.Values[k]
			}

			return a, nil
		},
		func(o interface{}) error {
			board := o.(*jira.Board)

			if board.Type != "scrum" {
				return nil
			}

			if re != nil {
				if !re.MatchString(board.Name) {
					return nil
				}
			}

			boardName := fmt.Sprintf("%s (%d)", ReplaceStringTrash(board.Name), board.ID)

			logrus.Infof("board %q", boardName)

			mdir := path.Join(s.config.Remote[s.remote].DestDir, "boards", boardName)

			err := os.MkdirAll(mdir, 0755)
			if err != nil {
				return err
			}

			err = s.sprints(mdir, board, refs)
			if err != nil {
				return err
			}

			err = s.backlog(mdir, board, refs)
			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	logrus.Infof("boards %d handled %d", count, handled)

	return nil
}
