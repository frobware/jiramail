package syncer

import (
	"fmt"
	"net/mail"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/legionus/jirasync/internal/cache"
	"github.com/legionus/jirasync/internal/client"
	"github.com/legionus/jirasync/internal/config"
	"github.com/legionus/jirasync/internal/jiraplus"
	"github.com/legionus/jirasync/internal/maildir"
	"github.com/legionus/jirasync/internal/message"
)

type JiraSyncer struct {
	client    *jiraplus.Client
	config    *config.Configuration
	remote    string
	usercache *cache.User
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
	s.usercache = cache.NewUserCache(s.client)
	return
}

func (s *JiraSyncer) writeMessage(mdir maildir.Dir, msg *mail.Message) error {
	messageID := message.HeaderID(msg)

	curMessageHash, err := getMessageHash(mdir, messageID)
	if err != nil {
		return err
	}

	d, err := mdir.NewDeliveryKey(messageID)
	if err != nil {
		return fmt.Errorf("can not create ongoing message delivery to the mailbox: %s", err)
	}

	newMessageHash, err := message.Write(d, msg)
	if err != nil {
		d.Abort()
		return err
	}

	if curMessageHash == newMessageHash {
		d.Abort()
		s.msgids[messageID] = struct{}{}
		return nil
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

	err = os.Remove(path.Join(string(mdir), "new", key))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	err = d.Close()
	if err != nil {
		return err
	}

	err = mdir.Purge(key)
	if err != nil {
		mailErr, ok := err.(*maildir.KeyError)
		if ok && mailErr.N != 0 {
			return err
		}
	}

	err = os.Link(
		filepath.Join(string(mdir), "new", key),
		filepath.Join(string(mdir), "cur", key+string(maildir.Separator)+"2,"+flags),
	)
	if err != nil {
		return fmt.Errorf("unable to link message %s from new to cur: %s", key, err)
	}

	err = os.Remove(filepath.Join(string(mdir), "new", key))
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
