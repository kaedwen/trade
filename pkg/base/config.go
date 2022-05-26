package base

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"
)

type ClientRequestID struct {
	SessionID string `json:"sessionId"`
	RequestID string `json:"requestId"`
}

type ConfigStatic struct {
	LogLevel      string `yaml:"log_level" default:"info"`
	ClientID      string `yaml:"client_id"`
	ClientSecret  string `yaml:"client_secret"`
	Zugangsnummer string `yaml:"zugangsnummer"`
	Pin           string `yaml:"pin"`
	Url           string `yaml:"url"`
	OAuthUrl      string `yaml:"oauth_url"`
	Influx        struct {
		Url    string `yaml:"url"`
		Token  string `yaml:"token"`
		Org    string `yaml:"org"`
		Bucket string `yaml:"bucket"`
	} `yaml:"influx"`
}

type ConfigRuntime struct {
	ConfigStatic          `yaml:",inline"`
	SessionID             string `yaml:"session_id"`
	RequestID             string `yaml:"request_id"`
	AccessToken           string `yaml:"access_token"`
	RefreshToken          string `yaml:"refresh_token"`
	SessionUUID           string `yaml:"sessionUUID"`
	ChallengeID           string `yaml:"challenge_id"`
	Tan                   string `yaml:"tan"`
	AccountUUID           string `yaml:"accountUUID"`
	DepotUUID             string `yaml:"depotUUID"`
	SettlementAccountUUID string `yaml:"settlementAccountUUID"`
	PostionUUID           string `yaml:"positionUUID"`
	DocumentUUID          string `yaml:"documentUUID"`
}

var Config ConfigRuntime
var RequestInfo struct {
	ClientRequestID ClientRequestID `json:"clientRequestId"`
}

func setupConfig() {

	if path, err := isConfigTraversePresent(); err == nil {
		loadConfig(path)
	} else if path, err := isConfigHomePresent(); err == nil {
		loadConfig(path)
	} else if path, err := isConfigEtcPresent(); err == nil {
		loadConfig(path)
	} else {
		log.Fatal().Msgf("no config provided")
	}

}

func isFilePresent(path string) (string, error) {
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else {
		return "", err
	}
}

func isConfigTraversePresent() (string, error) {

	path, err := os.Getwd()
	FatalIfError(err)

	for {
		target := filepath.Join(path, ".config.yaml")
		if file, err := isFilePresent(target); err == nil {
			return file, err
		} else {
			if path == "/" {
				return "", fmt.Errorf("no .config.yaml found up to /")
			}
			path = filepath.Dir(path)
		}
	}
}

func isConfigEtcPresent() (string, error) {
	return isFilePresent("/etc/trade/config.yaml")
}

func isConfigHomePresent() (string, error) {
	if dirname, err := os.UserHomeDir(); err == nil {
		return isFilePresent(filepath.Join(dirname, ".config/trade/config.yaml"))
	} else {
		return "", err
	}
}

func loadConfig(path string) {
	// set defaults
	Config.LogLevel = "debug"
	Config.Url = "https://api.comdirect.de/api"
	Config.OAuthUrl = "https://api.comdirect.de"

	configBytes, err := os.ReadFile(path)
	FatalIfError(err)

	err = yaml.Unmarshal(configBytes, &Config)
	FatalIfError(err)
}
