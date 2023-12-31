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
	DefaultEnvPath    = "config/config.env"
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
	ServerConfig ServiceConfig    `toml:"server"`
	DHTConfig    DHTServiceConfig `toml:"dht"`
}

type ServiceConfig struct {
	Environment Environment `toml:"env"`
	APIHost     string      `toml:"api_host"`
	APIPort     int         `toml:"api_port"`
	ListenPort  int         `toml:"listen_port"`
	BroadcastIP string      `toml:"broadcast_ip"`
	LogLocation string      `toml:"log_location"`
	LogLevel    string      `toml:"log_level"`
	DBFile      string      `toml:"db_file"`
}

type DHTServiceConfig struct {
	Name             string   `toml:"name"`
	Namespace        string   `toml:"namespace"`
	Topic            string   `toml:"topic"`
	LocalDiscovery   bool     `toml:"local_discovery"`
	ResolverEndpoint string   `toml:"resolver_endpoint"`
	BootstrapPeers   []string `toml:"bootstrap_peers"`
	// if set, the API will only accept signed messages
	EnforceSignedMessages bool `toml:"enforce_signed_messages"`
}

func GetDefaultConfig() Config {
	return Config{
		ServerConfig: ServiceConfig{
			Environment: EnvironmentDev,
			APIHost:     "0.0.0.0",
			APIPort:     8305,
			ListenPort:  8503,
			LogLocation: "log",
			LogLevel:    "debug",
			DBFile:      "diddht.db",
		},
		DHTConfig: DHTServiceConfig{
			Name:                  "tbd",
			Namespace:             "diddht",
			Topic:                 "diddht",
			LocalDiscovery:        true,
			ResolverEndpoint:      "https://dev.uniresolver.io/",
			EnforceSignedMessages: false,
		},
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
		logrus.Info("loading default config...")
		cfg = GetDefaultConfig()
	} else {
		if path == "" {
			logrus.Info("no config path provided, trying default config path...")
			path = DefaultConfigPath
		}
		if err = loadTOMLConfig(path, &cfg); err != nil {
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
	defaultConfig := false
	if path == "" {
		logrus.Info("no config path provided, loading default config...")
		defaultConfig = true
	} else if filepath.Ext(path) != Extension {
		return false, fmt.Errorf("file extension for path %q must be %q", path, Extension)
	}
	return defaultConfig, nil
}

func loadTOMLConfig(path string, cfg *Config) error {
	// load from TOML file
	if _, err := toml.DecodeFile(path, cfg); err != nil {
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
		cfg.DHTConfig.Name = nameEnv
	}

	broadcastIPEnv, present := os.LookupEnv(BroadcastIPEnvVar.String())
	if present {
		cfg.ServerConfig.BroadcastIP = broadcastIPEnv
	}
	return nil
}
