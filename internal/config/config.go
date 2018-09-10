package config

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type LogLevel struct {
	logrus.Level
}

func (d *LogLevel) UnmarshalText(data []byte) (err error) {
	d.Level, err = logrus.ParseLevel(strings.ToLower(string(data)))
	return
}

type Configuration struct {
	Core    Core
	SMTP    *SMTP
	Command Command
	Remote  map[string]*Remote
}

type Core struct {
	LogLevel   LogLevel
	LogFile    string
	LockDir    string
	SyncPeriod time.Duration
}

type SMTP struct {
	Addr         string
	Hostname     string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PlainAuth    bool
}

type Command struct {
	To string
}

type Remote struct {
	DestDir      string
	BaseURL      string
	Username     string
	Password     string
	ProjectMatch string
	BoardMatch   string
}
