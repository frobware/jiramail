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
			return NewEnvelope(cfg, c, from)
			//return &smtpd.BasicEnvelope{}, nil
		},
	}

	logrus.Infof("listen: %s", srv.Addr)

	return srv.ListenAndServe()
}
