package syncer

import (
	"io"
	"net/mail"
	"net/textproto"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/legionus/jiramail/internal/maildir"
	"github.com/legionus/jiramail/internal/message"
)

func tagDeletedMessage(f *os.File) error {
	m, err := mail.ReadMessage(f)
	if err != nil {
		return err
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	m.Header["Subject"] = []string{"[DELETED] " + m.Header.Get("Subject")}

	return message.Write(f, m)
}

func (s *JiraSyncer) CleanDir(mdir maildir.Dir) error {
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

		if s.config.Remote[s.remote].Delete == "tag" {
			fn, err := mdir.Filename(msgid)
			if err != nil {
				return err
			}
			f, err := os.OpenFile(fn, os.O_RDWR, 0644)
			if err != nil {
				return err
			}

			err = tagDeletedMessage(f)
			f.Close()

			if err != nil {
				return err
			}
		} else {
			err = mdir.Purge(msgid)
			if err != nil {
				return err
			}
		}
	}

	err = mdir.Clean()
	if err != nil {
		return err
	}
	return nil
}
