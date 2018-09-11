package smtp

import (
	"github.com/bradfitz/go-smtpd/smtpd"
	"github.com/sirupsen/logrus"

	"github.com/legionus/jiramail/internal/config"
)

func Server(cfg *config.Configuration) error {
	srv := &smtpd.Server{
		Addr:         cfg.SMTP.Addr,
		Hostname:     cfg.SMTP.Hostname,
		ReadTimeout:  cfg.SMTP.ReadTimeout,
		WriteTimeout: cfg.SMTP.WriteTimeout,
		PlainAuth:    cfg.SMTP.PlainAuth,
		OnNewMail: func(c smtpd.Connection, from smtpd.MailAddress) (smtpd.Envelope, error) {
			if cfg.SMTP.LogMessagesOnly {
				return &smtpd.BasicEnvelope{}, nil
			}
			return NewEnvelope(cfg, c, from)
		},
	}

	logrus.Infof("listen: %s", srv.Addr)

	return srv.ListenAndServe()
}
