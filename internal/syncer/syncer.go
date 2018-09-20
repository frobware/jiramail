package syncer

import (
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	"github.com/legionus/jiramail/internal/cache"
	"github.com/legionus/jiramail/internal/client"
	"github.com/legionus/jiramail/internal/config"
	"github.com/legionus/jiramail/internal/jiraconv"
	"github.com/legionus/jiramail/internal/jiraplus"
	"github.com/legionus/jiramail/internal/maildir"
	"github.com/legionus/jiramail/internal/message"
)

type JiraSyncer struct {
	client    *jiraplus.Client
	converter *jiraconv.Converter
	config    *config.Configuration
	remote    string
	msgids    map[string]struct{}
	projects  map[string]struct{}
}

func NewJiraSyncer(c *config.Configuration, remoteName string) (s *JiraSyncer, err error) {
	s = &JiraSyncer{
		config:   c,
		remote:   remoteName,
		msgids:   make(map[string]struct{}),
		projects: make(map[string]struct{}),
	}
	s.client, err = client.NewClient(s.config, s.remote)
	if err != nil {
		return
	}

	s.converter = jiraconv.NewConverter(s.remote, cache.NewUserCache(s.client))

	fields, _, err := s.client.Field.GetList()
	if err != nil {
		return
	}

	s.converter.SetJiraFields(fields)

	return
}

func (s *JiraSyncer) writeMessage(mdir maildir.Dir, msg *mail.Message) error {
	messageID := message.HeaderID(msg)

	curMessageHash, err := getMessageHash(mdir, messageID)
	if err != nil {
		return err
	}

	newMessageHash, err := message.MakeChecksum(msg)
	if err != nil {
		return err
	}

	if curMessageHash == newMessageHash {
		s.msgids[messageID] = struct{}{}
		return nil
	}

	d, err := mdir.NewDeliveryKey(messageID)
	if err != nil {
		return fmt.Errorf("can not create ongoing message delivery to the mailbox: %s", err)
	}

	msg.Header["X-Checksum"] = []string{newMessageHash}

	err = message.Write(d, msg)
	if err != nil {
		d.Abort()
		return err
	}

	if err = CloseDelivery(mdir, messageID, d); err != nil {
		return err
	}

	s.msgids[messageID] = struct{}{}
	return nil
}

func CloseDelivery(mdir maildir.Dir, key string, d *maildir.Delivery) error {
	flags, err := mdir.Flags(key)
	if err != nil {
		mailErr, ok := err.(*maildir.KeyError)
		if ok && mailErr.N != 0 {
			return err
		}
		flags = ""
	} else {
		flags = strings.Replace(flags, "S", "", -1)
	}

	err = mdir.Purge(key)
	if err != nil {
		mailErr, ok := err.(*maildir.KeyError)
		if ok && mailErr.N != 0 {
			return err
		}
	}

	err = d.Close()
	if err != nil {
		return err
	}

	return nil
}

func getMessageHash(mdir maildir.Dir, key string) (string, error) {
	fp, err := mdir.Filename(key)

	if err != nil {
		mailErr, ok := err.(*maildir.KeyError)
		if ok && mailErr.N == 0 {
			return "", nil
		}
		return "", err
	}

	return message.GetChecksum(fp)
}

func Maildir(p string) (maildir.Dir, error) {
	st, err := os.Stat(p)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		err := os.MkdirAll(filepath.Dir(p), 0755)
		if err != nil {
			return "", err
		}
		if err := maildir.Dir(p).Create(); err != nil {
			return "", fmt.Errorf("unable to create maildir: %s", err)
		}
	} else {
		if !st.Mode().IsDir() {
			return "", fmt.Errorf("dirctory expected: %s", p)
		}
	}
	return maildir.Dir(p), nil
}
