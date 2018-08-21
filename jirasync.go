package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/legionus/getopt"
	"github.com/legionus/jirasync/internal/config"
	"github.com/legionus/jirasync/internal/smtp"
	"github.com/legionus/jirasync/internal/syncer"
)

var (
	prog           = ""
	version        = "1.0"
	configcile     = ""
	showVersionOpt = false
	showHelpOpt    = false
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

func main() {
	prog = filepath.Base(os.Args[0])
	oneSync := false
	configFile := os.Getenv("HOME") + "/.jirasyncrc"

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

	if len(cfg.Core.LogFile) > 0 {
		f, err := os.OpenFile(cfg.Core.LogFile, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			logrus.Fatalf("%s", err)
		}
		defer f.Close()
		logrus.SetOutput(f)
	}

	logrus.SetLevel(cfg.Core.LogLevel.Level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
	})

	var wg sync.WaitGroup

	wg.Add(1)
	go func(c *config.Configuration) {
		defer wg.Done()

		ticker := time.NewTicker(c.Core.SyncPeriod)
		defer ticker.Stop()

		for {
			for name := range c.Remote {
				logrus.Infof("syncing %s", name)

				jiraSyncer, err := syncer.NewJiraSyncer(c, name)
				if err != nil {
					logrus.Fatalf("%s", err)
				}

				err = jiraSyncer.Boards()
				if err != nil {
					logrus.Fatalf("%s", err)
				}
				logrus.Infof("synchronization of boards for %s is completed", name)

				err = jiraSyncer.Projects()
				if err != nil {
					logrus.Fatalf("%s", err)
				}
				logrus.Infof("synchronization of projects for %s is completed", name)

			}
			logrus.Infof("synchronization is completed")

			if oneSync {
				break
			}

			select {
			case <-ticker.C:
				logrus.Info("re-sync")
			}
		}
	}(cfg)

	if !oneSync && cfg.SMTP != nil {
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
