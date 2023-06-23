package config

const (
	EnvironmentDev  Environment = "dev"
	EnvironmentTest Environment = "test"
	EnvironmentProd Environment = "prod"
)

type (
	Environment         string
	EnvironmentVariable string
)

func (e EnvironmentVariable) String() string {
	return string(e)
}

type Config struct {
	Environment    Environment `toml:"env" conf:"default:dev"`
	APIHost        string      `toml:"api_host" conf:"default:0.0.0.0:8305"`
	LogLocation    string      `toml:"log_location" conf:"default:log"`
	LogLevel       string      `toml:"log_level" conf:"default:debug"`
	DBFile         string      `toml:"db_file" conf:"default:diddht.db"`
	Name           string      `toml:"name" conf:"default:diddht-og"`
	Namespace      string      `toml:"namespace" conf:"default:diddht"`
	Topic          string      `toml:"topic" conf:"default:diddht"`
	LocalDiscovery bool        `toml:"local_discovery" conf:"default:false"`
}

func GetDefaultConfig() Config {
	return Config{
		Environment:    EnvironmentDev,
		APIHost:        "0.0.0.0:8305",
		LogLocation:    "log",
		LogLevel:       "debug",
		DBFile:         "diddht.db",
		Name:           "diddht-og",
		Namespace:      "diddht",
		Topic:          "diddht",
		LocalDiscovery: true,
	}
}
