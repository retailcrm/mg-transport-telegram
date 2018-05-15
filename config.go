package main

import (
	"io/ioutil"
	"path/filepath"

	logging "github.com/op/go-logging"
	yaml "gopkg.in/yaml.v2"
)

// TransportConfig struct
type TransportConfig struct {
	LogLevel   logging.Level    `yaml:"log_level"`
	Database   DatabaseConfig   `yaml:"database"`
	SentryDSN  string           `yaml:"sentry_dsn"`
	HttpServer HttpServerConfig `yaml:"http_server"`
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

type HttpServerConfig struct {
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
