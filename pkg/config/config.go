package config

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

var ErrNoConfigAvailable = errors.New("no config available")

type Config struct {
	ApiAddress   *URL   `yaml:"apiAddress"`
	TokenAddress *URL   `yaml:"tokenAddress"`
	ClientId     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
	AccountId    string `yaml:"accountId"`
	Pin          string `yaml:"pin"`
}

func NewConfig() (*Config, error) {
	p := []string{"/etc/trade"}

	if h, err := os.UserHomeDir(); err == nil {
		p = append(p, filepath.Join(h, ".config", "trade"))
	}

	p = append(p, "./")

	slices.Reverse(p)

	for _, p := range p {
		cfg, err := newConfigFromFile(path.Join(p, "config.yaml"))
		if err == nil {
			return cfg, nil
		}
	}

	return nil, ErrNoConfigAvailable
}

func newConfigFromFile(f string) (*Config, error) {
	data, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
