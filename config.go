package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/op/go-logging"
	"gopkg.in/yaml.v2"
)

// TransportConfig struct
type TransportConfig struct {
	LogLevel       logging.Level    `yaml:"log_level"`
	Database       DatabaseConfig   `yaml:"database"`
	SentryDSN      string           `yaml:"sentry_dsn"`
	HTTPServer     HTTPServerConfig `yaml:"http_server"`
	Debug          bool             `yaml:"debug"`
	UpdateInterval int              `yaml:"update_interval"`
	ConfigAWS      ConfigAWS        `yaml:"config_aws"`
	Credentials    []string         `yaml:"credentials"`
	TransportInfo  TransportInfo    `yaml:"transport_info"`
}

type TransportInfo struct {
	Name     string `yaml:"name"`
	Code     string `yaml:"code"`
	LogoPath string `yaml:"logo_path"`
}

// ConfigAWS struct
type ConfigAWS struct {
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region          string `yaml:"region"`
	Bucket          string `yaml:"bucket"`
	FolderName      string `yaml:"folder_name"`
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
