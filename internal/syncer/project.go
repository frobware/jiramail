package syncer

import (
	"fmt"
	"path"
	"regexp"

	"github.com/andygrunwald/go-jira"
	"github.com/sirupsen/logrus"

	"github.com/legionus/jiramail/internal/jiraconv"
	"github.com/legionus/jiramail/internal/maildir"
)

func (s *JiraSyncer) project(mdir maildir.Dir, projectKey string) error {
	project, _, err := s.client.Project.Get(projectKey)
	if err != nil {
		return fmt.Errorf("unable to get project: %s: %s", projectKey, err)
	}

	logrus.Infof("project %s", project.Key)

	refs := []string{jiraconv.RemoteMessageID(s.remote)}

	msg, err := jiraconv.NewConverter(s.remote, s.usercache).Project(project, refs)
	if err != nil {
		return err
	}

	err = s.writeMessage(mdir, msg)
	if err != nil {
		return err
	}

	err = s.issues(mdir, fmt.Sprintf("project = %s", projectKey), append(refs, msg.Header.Get("Message-ID")))
	if err != nil {
		return err
	}

	return nil
}

func (s *JiraSyncer) Projects() error {
	var (
		re *regexp.Regexp
	)

	if len(s.config.Remote[s.remote].ProjectMatch) > 0 {
		re = regexp.MustCompile(s.config.Remote[s.remote].ProjectMatch)
	}

	projectList, _, err := s.client.Project.ListWithOptions(&jira.GetQueryOptions{
		Expand: "description,lead",
		Fields: "*all",
	})

	if err != nil {
		return fmt.Errorf("unable to get projects: %s", err)
	}

	for _, project := range *projectList {
		if re != nil {
			_, ok := s.projects[project.Key]

			if !ok && !re.MatchString(project.Key) {
				continue
			}
		}

		mdir, err := Maildir(path.Join(s.config.Remote[s.remote].DestDir, "projects", project.Key))
		if err != nil {
			return err
		}

		err = s.project(mdir, project.Key)
		if err != nil {
			return err
		}

		// Garbage collection
		err = s.CleanDir(mdir)
		if err != nil {
			return err
		}
	}

	return nil
}
