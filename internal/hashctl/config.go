package hashctl

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type envLookup func(string) string

type Config struct {
	APIURL           string        `json:"api_url"`
	ConfigPath       string        `json:"-"`
	Output           string        `json:"output"`
	Timeout          time.Duration `json:"-"`
	CorrelationID    string        `json:"-"`
	LocalPrincipalID string        `json:"-"`
	LocalGroups      []string      `json:"-"`
	LocalAppRoles    []string      `json:"-"`
	BearerToken      string        `json:"-"`
	BearerTokenFile  string        `json:"-"`
	PollInterval     time.Duration `json:"-"`
	PollTimeout      time.Duration `json:"-"`
}

func defaultConfig() Config {
	return Config{
		Output:       "human",
		Timeout:      30 * time.Second,
		PollInterval: defaultPollInterval,
		PollTimeout:  defaultPollTimeout,
	}
}

func resolveConfig(flags Config, getenv envLookup) (Config, error) {
	cfg := defaultConfig()
	cfg.ConfigPath = flags.ConfigPath

	if cfg.ConfigPath == "" {
		cfg.ConfigPath = getenv("HASHCTL_CONFIG")
	}
	if cfg.ConfigPath == "" {
		if dir, err := os.UserConfigDir(); err == nil && dir != "" {
			cfg.ConfigPath = filepath.Join(dir, "hashctl", "config.json")
		}
	}
	if cfg.ConfigPath != "" {
		fileCfg, err := readConfigFile(cfg.ConfigPath)
		if err != nil {
			return Config{}, err
		}
		if fileCfg.APIURL != "" {
			cfg.APIURL = fileCfg.APIURL
		}
		if fileCfg.Output != "" {
			cfg.Output = fileCfg.Output
		}
	}

	if envAPI := getenv("HASH_ENGINE_API"); envAPI != "" {
		cfg.APIURL = envAPI
	}
	if flags.APIURL != "" {
		cfg.APIURL = flags.APIURL
	}
	if flags.Output != "" {
		cfg.Output = flags.Output
	}
	if flags.Timeout != 0 {
		cfg.Timeout = flags.Timeout
	}
	if flags.CorrelationID != "" {
		cfg.CorrelationID = flags.CorrelationID
	}
	if flags.LocalPrincipalID != "" {
		cfg.LocalPrincipalID = flags.LocalPrincipalID
	}
	cfg.LocalGroups = append(cfg.LocalGroups, flags.LocalGroups...)
	cfg.LocalAppRoles = append(cfg.LocalAppRoles, flags.LocalAppRoles...)
	if token := strings.TrimSpace(getenv("HASH_ENGINE_BEARER_TOKEN")); token != "" {
		cfg.BearerToken = token
	}
	if flags.BearerTokenFile != "" {
		cfg.BearerTokenFile = flags.BearerTokenFile
		token, err := os.ReadFile(flags.BearerTokenFile)
		if err != nil {
			return Config{}, fmt.Errorf("read bearer token file: %w", err)
		}
		cfg.BearerToken = strings.TrimSpace(string(token))
	}
	if flags.PollInterval != 0 {
		cfg.PollInterval = flags.PollInterval
	}
	if flags.PollTimeout != 0 {
		cfg.PollTimeout = flags.PollTimeout
	}

	if cfg.Output != "human" && cfg.Output != "json" {
		return Config{}, fmt.Errorf("output must be human or json")
	}
	if cfg.APIURL == "" {
		return Config{}, fmt.Errorf("API URL is required; set --api-url, HASH_ENGINE_API, or hashctl config.json")
	}
	parsed, err := url.Parse(cfg.APIURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return Config{}, fmt.Errorf("API URL must be an absolute URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return Config{}, fmt.Errorf("API URL scheme must be http or https")
	}
	if cfg.Timeout <= 0 {
		return Config{}, fmt.Errorf("timeout must be greater than zero")
	}
	if cfg.PollInterval <= 0 {
		return Config{}, fmt.Errorf("poll interval must be greater than zero")
	}
	if cfg.PollTimeout <= 0 {
		return Config{}, fmt.Errorf("poll timeout must be greater than zero")
	}
	return cfg, nil
}

func readConfigFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config file: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file: %w", err)
	}
	return cfg, nil
}
