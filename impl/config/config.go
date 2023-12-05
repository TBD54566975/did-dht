package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	ConfigPath EnvironmentVariable = "CONFIG_PATH"
	// BootstrapPeers A comma-separated list of bootstrap peers to connect to on startup.
	BootstrapPeers EnvironmentVariable = "BOOTSTRAP_PEERS"
)

type (
	Environment         string
	EnvironmentVariable string
)

func (e EnvironmentVariable) String() string {
	return string(e)
}

type Config struct {
	ServerConfig ServerConfig       `toml:"server"`
	DHTConfig    DHTServiceConfig   `toml:"dht"`
	PkarrConfig  PKARRServiceConfig `toml:"pkarr"`
}

type ServerConfig struct {
	Environment Environment `toml:"env"`
	APIHost     string      `toml:"api_host"`
	APIPort     int         `toml:"api_port"`
	BaseURL     string      `toml:"base_url"`
	LogLocation string      `toml:"log_location"`
	LogLevel    string      `toml:"log_level"`
	DBFile      string      `toml:"db_file"`
}

type DHTServiceConfig struct {
	BootstrapPeers []string `toml:"bootstrap_peers"`
}

type PKARRServiceConfig struct {
	RepublishCRON    string `toml:"republish_cron"`
	CacheTTLSeconds  int    `toml:"cache_ttl_seconds"`
	CacheSizeLimitMB int    `toml:"cache_size_limit_mb"`
}

func GetDefaultConfig() Config {
	return Config{
		ServerConfig: ServerConfig{
			Environment: EnvironmentDev,
			APIHost:     "0.0.0.0",
			APIPort:     8305,
			BaseURL:     "http://localhost:8305",
			LogLocation: "log",
			LogLevel:    "debug",
			DBFile:      "diddht.db",
		},
		DHTConfig: DHTServiceConfig{
			BootstrapPeers: GetDefaultBootstrapPeers(),
		},
		PkarrConfig: PKARRServiceConfig{
			RepublishCRON:    "0 */2 * * *",
			CacheTTLSeconds:  600,
			CacheSizeLimitMB: 2000,
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

	bootstrapPeers, present := os.LookupEnv(BootstrapPeers.String())
	if present {
		cfg.DHTConfig.BootstrapPeers = strings.Split(bootstrapPeers, ",")
	}
	return nil
}

// GetDefaultBootstrapPeers returns a list of default bootstrap peers for the DHT.
func GetDefaultBootstrapPeers() []string {
	return []string{
		"router.magnets.im:6881",
		"router.bittorrent.com:6881",
		"dht.transmissionbt.com:6881",
		"router.utorrent.com:6881",
		"router.nuh.dev:6881",
	}
}
