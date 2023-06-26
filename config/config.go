package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	DefaultConfigPath = "config/config.toml"
	DefaultEnvPath    = "config/.env"
	Extension         = ".toml"

	EnvironmentDev  Environment = "dev"
	EnvironmentTest Environment = "test"
	EnvironmentProd Environment = "prod"

	ConfigPath        EnvironmentVariable = "CONFIG_PATH"
	NameEnvVar        EnvironmentVariable = "NAME"
	BroadcastIPEnvVar EnvironmentVariable = "BROADCAST_IP"
)

type (
	Environment         string
	EnvironmentVariable string
)

func (e EnvironmentVariable) String() string {
	return string(e)
}

type Config struct {
	Environment    Environment `toml:"env"`
	APIHost        string      `toml:"api_host"`
	APIPort        int         `toml:"api_port"`
	ListenPort     int         `toml:"listen_port"`
	BroadcastIP    string      `toml:"broadcast_ip"`
	BootstrapPeers []string    `toml:"bootstrap_peers"`
	LogLocation    string      `toml:"log_location"`
	LogLevel       string      `toml:"log_level"`
	DBFile         string      `toml:"db_file" `
	Name           string      `toml:"name"`
	Namespace      string      `toml:"namespace"`
	Topic          string      `toml:"topic"`
	LocalDiscovery bool        `toml:"local_discovery"`
}

func GetDefaultConfig() Config {
	return Config{
		Environment:    EnvironmentDev,
		APIHost:        "0.0.0.0",
		APIPort:        8305,
		ListenPort:     8503,
		BootstrapPeers: []string{"/ip4/54.226.19.143/tcp/8503/p2p/12D3KooWEgKRaVKRpdzsYDgFtXVGYGC21o5BrxMivvpXtapgcAXm"},
		LogLocation:    "log",
		LogLevel:       "debug",
		DBFile:         "diddht.db",
		Name:           "gabe",
		Namespace:      "diddht",
		Topic:          "diddht",
		LocalDiscovery: true,
	}
}

// LoadConfig attempts to load a TOML config file from the given path, and coerce it into our object model.
// Before loading, defaults are applied on certain properties, which are overwritten if specified in the TOML file.
func LoadConfig(path string) (*Config, error) {
	loadDefaultConfig, err := checkValidConfigPath(path)
	if err != nil {
		return nil, errors.Wrap(err, "validating config path")
	}

	var cfg Config
	if loadDefaultConfig {
		cfg = GetDefaultConfig()
	} else {
		if path == "" {
			logrus.Info("no config path provided, trying default config path...")
			path = DefaultConfigPath
		}
		if err = loadTOMLConfig(path, cfg); err != nil {
			return nil, errors.Wrap(err, "load toml config")
		}
	}

	if err = applyEnvVariables(cfg); err != nil {
		return nil, errors.Wrap(err, "apply env variables")
	}
	return &cfg, nil
}

func checkValidConfigPath(path string) (bool, error) {
	// no path, load default config
	if path == "" {
		logrus.Info("no config path provided, loading default config...")
		return true, nil
	}
	if filepath.Ext(path) != Extension {
		return false, fmt.Errorf("file extension for path %q must be %q", path, Extension)
	}
	return true, nil
}

func loadTOMLConfig(path string, cfg Config) error {
	// load from TOML file
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return errors.Wrapf(err, "could not load config: %s", path)
	}
	return nil
}

func applyEnvVariables(cfg Config) error {
	if err := godotenv.Load(DefaultEnvPath); err != nil {
		// The error indicates that the file or directory does not exist.
		if os.IsNotExist(err) {
			logrus.Info("no .env file found, skipping apply env variables...")
		} else {
			return errors.Wrap(err, "dotenv parsing")
		}
	}

	nameEnv, present := os.LookupEnv(NameEnvVar.String())
	if present {
		cfg.Name = nameEnv
	}

	broadcastIPEnv, present := os.LookupEnv(BroadcastIPEnvVar.String())
	if present {
		cfg.BroadcastIP = broadcastIPEnv
	}
	return nil
}
