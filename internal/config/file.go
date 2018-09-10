package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

func substHome(s string) string {
	arr := strings.Split(s, string(os.PathSeparator))
	if arr[0] == "~" {
		arr[0] = os.Getenv("HOME")
		return filepath.Join(arr...)
	}
	return s
}

func Read(filename string) (*Configuration, error) {
	cfg := &Configuration{}

	cfg.Core.SyncPeriod = 5 * time.Minute

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = yaml.NewDecoder(f).Decode(cfg)
	if err != nil {
		return nil, err
	}

	cfg.Core.LogFile = substHome(cfg.Core.LogFile)
	cfg.Core.LockDir = substHome(cfg.Core.LockDir)
	for name := range cfg.Remote {
		cfg.Remote[name].DestDir = substHome(cfg.Remote[name].DestDir)
	}

	return cfg, nil
}
