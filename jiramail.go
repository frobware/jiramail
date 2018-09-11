package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
	"github.com/sirupsen/logrus"

	"github.com/legionus/getopt"
	"github.com/legionus/jiramail/internal/config"
	"github.com/legionus/jiramail/internal/smtp"
	"github.com/legionus/jiramail/internal/syncer"
)

var (
	prog           = ""
	version        = "1.0"
	showVersionOpt = false
	showHelpOpt    = false
	lockDir        = ""
	exitCode       = 0
)

func usage() {
	fmt.Fprintf(os.Stdout, `
Usage: %[1]s [options]

This utility is a simple interface to jira.

Options:
  -1, --onesync          run synchronization once and exit;
  -c, --config=FILE      use an alternative configuration file;
  -V, --version          print program version and exit;
  -h, --help             show this text and exit.

Report bugs to author.

`,
		prog)
	os.Exit(0)
}

func showUsage(*getopt.Option, getopt.NameType, string) error {
	usage()
	return nil
}

func showVersion(*getopt.Option, getopt.NameType, string) error {
	fmt.Fprintf(os.Stdout, `%s version %s
Written by Alexey Gladkov.

Copyright (C) 2018  Alexey Gladkov <gladkov.alexey@gmail.com>
This is free software; see the source for copying conditions.  There is NO
warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

`,
		prog, version)
	os.Exit(0)
	return nil
}

func defaultSigHandler(sig os.Signal) error {
	logrus.Infof("got signal %s, exit", sig)
	if len(lockDir) > 0 {
		_ = os.RemoveAll(lockDir)
	}
	os.Exit(exitCode)
	return daemon.ErrStop
}

func syncJira(c *config.Configuration) error {
	for name := range c.Remote {
		jiraSyncer, err := syncer.NewJiraSyncer(c, name)
		if err != nil {
			return err
		}

		err = jiraSyncer.Boards()
		if err != nil {
			return err
		}

		err = jiraSyncer.Projects()
		if err != nil {
			return err
		}
	}

	logrus.Infof("synchronization is completed")
	return nil
}

func main() {
	prog = filepath.Base(os.Args[0])
	oneSync := false
	configFile := os.Getenv("HOME") + "/.jiramailrc"

	opts := &getopt.Getopt{
		AllowAbbrev: true,
		Options: []getopt.Option{
			{'h', "help", getopt.NoArgument,
				showUsage,
			},
			{'V', "version", getopt.NoArgument,
				showVersion,
			},
			{'c', "config", getopt.RequiredArgument,
				func(o *getopt.Option, t getopt.NameType, v string) (err error) {
					info, err := os.Stat(v)
					if err != nil {
						return
					}
					if !info.Mode().IsRegular() {
						return fmt.Errorf("regular file required")
					}
					configFile = v
					return
				},
			},
			{'1', "onesync", getopt.NoArgument,
				func(o *getopt.Option, t getopt.NameType, v string) (err error) {
					oneSync = true
					return
				},
			},
		},
	}

	prog = filepath.Base(os.Args[0])
	if err := opts.Parse(os.Args); err != nil {
		logrus.Fatalf("%v", err)
	}

	cfg, err := config.Read(configFile)
	if err != nil {
		logrus.Fatalf("%s", err)
	}

	logrus.SetLevel(cfg.Core.LogLevel.Level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
	})

	if oneSync {
		err = syncJira(cfg)
		if err != nil {
			logrus.Fatalf("%s", err)
		}
		return
	}

	if len(cfg.Core.LockDir) > 0 {
		err = os.Mkdir(cfg.Core.LockDir, 0700)
		if err != nil {
			if !os.IsExist(err) {
				logrus.Fatalf("%s", err)
			}
			return
		}
		defer func() {
			_ = os.RemoveAll(cfg.Core.LockDir)
		}()
		lockDir = cfg.Core.LockDir
	}

	if len(cfg.Core.LogFile) > 0 {
		f, err := os.OpenFile(cfg.Core.LogFile, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			logrus.Fatalf("%s", err)
		}
		defer f.Close()
		logrus.SetOutput(f)
	}

	daemon.SetSigHandler(defaultSigHandler, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)

	var wg sync.WaitGroup

	wg.Add(1)
	go func(c *config.Configuration) {
		defer wg.Done()
		err := daemon.ServeSignals()
		if err != nil {
			logrus.Errorf("%s", err)
		}
	}(cfg)

	wg.Add(1)
	go func(c *config.Configuration) {
		defer wg.Done()

		ticker := time.NewTicker(c.Core.SyncPeriod)
		defer ticker.Stop()

		for {
			err := syncJira(c)
			if err != nil {
				logrus.Fatalf("%s", err)
			}
			<-ticker.C
		}
	}(cfg)

	if cfg.SMTP != nil {
		wg.Add(1)
		go func(c *config.Configuration) {
			defer wg.Done()

			for {
				err := smtp.Server(c)
				if err != nil {
					logrus.Errorf("%s", err)
				}
				time.Sleep(3 * time.Second)
			}
		}(cfg)
	}

	wg.Wait()

	logrus.Infof("bye!")
}
