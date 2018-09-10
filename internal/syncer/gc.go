package syncer

import (
	"net/textproto"

	"github.com/sirupsen/logrus"

	"github.com/legionus/jiramail/internal/maildir"
	"github.com/legionus/jiramail/internal/message"
)

func (s *JiraSyncer) CleanDir(mdir maildir.Dir) error {
	_, err := mdir.Unseen()
	if err != nil {
		return err
	}

	msgids, err := mdir.Keys()
	if err != nil {
		return err
	}

	for _, msgid := range msgids {
		if _, ok := s.msgids[msgid]; ok {
			continue
		}

		headers := make(textproto.MIMEHeader)

		err = message.DecodeMessageID(msgid, headers)
		if err != nil {
			logrus.Warnf("unable to decode MessageID %q: %s", msgid, err)
		}

		logrus.Infof("removing obsolete message %s (%#+v)", msgid, headers)

		err = mdir.Purge(msgid)
		if err != nil {
			return err
		}
	}

	err = mdir.Clean()
	if err != nil {
		return err
	}
	return nil
}
