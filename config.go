package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/op/go-logging"
	"gopkg.in/yaml.v2"
)

// TransportConfig struct
type TransportConfig struct {
	LogLevel   logging.Level    `yaml:"log_level"`
	Database   DatabaseConfig   `yaml:"database"`
	SentryDSN  string           `yaml:"sentry_dsn"`
	HTTPServer HTTPServerConfig `yaml:"http_server"`
	Debug      bool             `yaml:"debug"`
	UpdateTime int              `yaml:"update_time"`
	ConfigAWS  ConfigAWS        `yaml:"config_aws"`
}

// ConfigAWS struct
type ConfigAWS struct {
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region          string `yaml:"region"`
	Bucket          string `yaml:"bucket"`
	ContentType     string `yaml:"content_type"`
}

// DatabaseConfig struct
type DatabaseConfig struct {
	Connection         string `yaml:"connection"`
	Logging            bool   `yaml:"logging"`
	TablePrefix        string `yaml:"table_prefix"`
	MaxOpenConnections int    `yaml:"max_open_connections"`
	MaxIdleConnections int    `yaml:"max_idle_connections"`
	ConnectionLifetime int    `yaml:"connection_lifetime"`
}

// HTTPServerConfig struct
type HTTPServerConfig struct {
	Host   string `yaml:"host"`
	Listen string `yaml:"listen"`
}

// LoadConfig read configuration file
func LoadConfig(path string) *TransportConfig {
	var err error

	path, err = filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	source, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var config TransportConfig
	if err = yaml.Unmarshal(source, &config); err != nil {
		panic(err)
	}

	return &config
}
