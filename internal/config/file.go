package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

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

	return cfg, nil
}
