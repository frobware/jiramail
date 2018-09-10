package smtp

import (
	"bufio"
	"fmt"
	"net/mail"
	"net/textproto"
	"strings"

	"github.com/bradfitz/go-smtpd/smtpd"
	"github.com/sirupsen/logrus"

	"github.com/legionus/jiramail/internal/config"
	"github.com/legionus/jiramail/internal/message"
	"github.com/legionus/jiramail/internal/smtp/command"
	"github.com/legionus/jiramail/internal/smtp/command/factory"

	_ "github.com/legionus/jiramail/internal/smtp/command/directive"
	_ "github.com/legionus/jiramail/internal/smtp/command/replace"
	_ "github.com/legionus/jiramail/internal/smtp/command/reply"
)

type envelope struct {
	from   smtpd.MailAddress
	rcpts  []smtpd.MailAddress
	config *config.Configuration
	buf    *strings.Builder
}

func NewEnvelope(cfg *config.Configuration, c smtpd.Connection, from smtpd.MailAddress) (smtpd.Envelope, error) {
	e := &envelope{
		config: cfg,
		from:   from,
	}
	return e, nil
}

func (e *envelope) getHandler(header textproto.MIMEHeader, key string) (command.Handler, error) {
	hdr := header.Get(key)
	if hdr == "" {
		return nil, mail.ErrHeaderNotPresent
	}

	list, err := mail.ParseAddressList(hdr)

	if err != nil {
		return nil, err
	}

	for _, addr := range list {
		handler, err := factory.Get(addr.Address)
		if err != nil {
			if err != factory.InvalidMailHandlerError {
				return nil, err
			}
		} else {
			return handler, nil
		}
	}

	return nil, factory.InvalidMailHandlerError
}

func (e *envelope) AddRecipient(rcpt smtpd.MailAddress) error {
	e.rcpts = append(e.rcpts, rcpt)
	return nil
}

func (e *envelope) BeginData() error {
	if len(e.rcpts) == 0 {
		return smtpd.SMTPError("554 5.5.1 Error: no valid recipients")
	}
	e.buf = &strings.Builder{}
	return nil
}

func (e *envelope) Write(line []byte) error {
	if e.buf == nil {
		return nil
	}
	_, err := e.buf.WriteString(string(line))
	return err
}

func (e *envelope) validate(header textproto.MIMEHeader) error {
	for _, name := range []string{"Message-ID", "References", "In-Reply-To"} {
		v := header.Get(name)
		if len(v) == 0 {
			return fmt.Errorf("header %s is empty", name)
		}
	}
	return nil
}

func (e *envelope) getMessage() (*command.Mail, error) {
	var err error

	msg := &command.Mail{}

	tp := textproto.NewReader(bufio.NewReader(strings.NewReader(e.buf.String())))

	msg.Header, err = tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(tp.R)

	for scanner.Scan() {
		msg.Body = append(msg.Body, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return msg, nil
}

func (e *envelope) Close() error {
	if e.buf == nil {
		return nil
	}

	msg, err := e.getMessage()
	if err != nil {
		return smtpd.SMTPError("550 " + err.Error())
	}

	if err = e.validate(msg.Header); err != nil {
		return smtpd.SMTPError("550 " + err.Error())
	}

	for _, ref := range strings.Fields(msg.Header["References"][0]) {
		if !strings.HasSuffix(ref, ".jira>") {
			continue
		}
		if err = message.DecodeMessageID(ref, msg.Header); err != nil {
			return smtpd.SMTPError("550 " + err.Error())
		}
	}

	var handler command.Handler

	cmds := command.GetJiraBlock(msg)
	if len(cmds) > 0 {
		handler, err = factory.Get("bot@jira")
		if err != nil {
			return smtpd.SMTPError("550 " + err.Error())
		}
		err = handler.Handle(e.config, &command.Mail{
			Header: msg.Header,
			Body:   cmds,
		})
		if err != nil {
			return smtpd.SMTPError("550 " + err.Error())
		}
	}

	for _, key := range []string{"To", "Cc", "Bcc"} {
		handler, err = e.getHandler(msg.Header, key)
		if err == nil {
			break
		}
		if err == factory.InvalidMailHandlerError {
			continue
		}
		return smtpd.SMTPError("550 " + err.Error())
	}

	if err == factory.InvalidMailHandlerError {
		handler, err = factory.Get("reply@jira")
		if err != nil {
			return smtpd.SMTPError("550 " + err.Error())
		}
	}

	err = handler.Handle(e.config, msg)
	if err == nil {
		return nil
	}

	logrus.Errorf("unable to handle message: %s", err)

	return smtpd.SMTPError("550 " + err.Error())
}
