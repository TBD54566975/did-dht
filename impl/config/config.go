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
	ServiceName       = "did-dht"
	DefaultConfigPath = "config/config.toml"
	DefaultEnvPath    = "config/config.env"
	Extension         = ".toml"

	EnvironmentDev  Environment = "dev"
	EnvironmentTest Environment = "test"
	EnvironmentProd Environment = "prod"

	ConfigPath EnvironmentVariable = "CONFIG_PATH"
	// BootstrapPeers A comma-separated list of bootstrap peers to connect to on startup.
	BootstrapPeers EnvironmentVariable = "BOOTSTRAP_PEERS"
	StorageURI     EnvironmentVariable = "STORAGE_URI"
	LogLevel       EnvironmentVariable = "LOG_LEVEL"
)

var Version = "devel"

type (
	Environment         string
	EnvironmentVariable string
)

func (e EnvironmentVariable) String() string {
	return string(e)
}

type Config struct {
	Log          LogConfig          `toml:"log"`
	ServerConfig ServerConfig       `toml:"server"`
	DHTConfig    DHTServiceConfig   `toml:"dht"`
	PkarrConfig  PkarrServiceConfig `toml:"pkarr"`
}

type ServerConfig struct {
	Environment Environment `toml:"env"`
	APIHost     string      `toml:"api_host"`
	APIPort     int         `toml:"api_port"`
	BaseURL     string      `toml:"base_url"`
	StorageURI  string      `toml:"storage_uri"`
	Telemetry   bool        `toml:"telemetry"`
}

type DHTServiceConfig struct {
	BootstrapPeers []string `toml:"bootstrap_peers"`
}

type PkarrServiceConfig struct {
	RepublishCRON    string `toml:"republish_cron"`
	CacheTTLSeconds  int    `toml:"cache_ttl_seconds"`
	CacheSizeLimitMB int    `toml:"cache_size_limit_mb"`
}

type LogConfig struct {
	Level string `toml:"level"`
}

func GetDefaultConfig() Config {
	return Config{
		ServerConfig: ServerConfig{
			Environment: EnvironmentDev,
			APIHost:     "0.0.0.0",
			APIPort:     8305,
			BaseURL:     "http://localhost:8305",
			StorageURI:  "bolt://diddht.db",
			Telemetry:   false,
		},
		DHTConfig: DHTServiceConfig{
			BootstrapPeers: GetDefaultBootstrapPeers(),
		},
		PkarrConfig: PkarrServiceConfig{
			RepublishCRON:    "0 */2 * * *",
			CacheTTLSeconds:  600,
			CacheSizeLimitMB: 500,
		},
		Log: LogConfig{
			Level: logrus.InfoLevel.String(),
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

	if err = applyEnvVariables(&cfg); err != nil {
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

func applyEnvVariables(cfg *Config) error {
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

	storage, present := os.LookupEnv(StorageURI.String())
	if present {
		cfg.ServerConfig.StorageURI = storage
	}

	levelString, present := os.LookupEnv(LogLevel.String())
	if present {
		_, err := logrus.ParseLevel(levelString)
		if err != nil {
			logrus.WithField("requested_level", levelString).Warn("unable to parse log level requested in environment variable")
		} else {
			cfg.Log.Level = levelString
		}
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
