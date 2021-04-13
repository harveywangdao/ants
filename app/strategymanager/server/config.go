package server

import (
	"encoding/json"
	"io/ioutil"

	"github.com/harveywangdao/ants/app/scheduler/util/logger"

	"gopkg.in/yaml.v2"
)

type LogConfig struct {
	LogPath  string `yaml:"logPath" json:"logPath"`
	LogLevel string `yaml:"logLevel" json:"logLevel"`
}

type EtcdConfig struct {
	Endpoints []string `yaml:"endpoints" json:"endpoints"`
}

type GrpcServerConfig struct {
	Name string `yaml:"name" json:"name"`
	Port string `yaml:"port" json:"port"`
}

type HttpServerConfig struct {
	Port string `yaml:"port" json:"port"`
}

type DatabaseConfig struct {
	Address    string `yaml:"address" json:"address"`
	Username   string `yaml:"username" json:"username"`
	Password   string `yaml:"password" json:"password"`
	DriverName string `yaml:"driverName" json:"driverName"`
	DbName     string `yaml:"dbName" json:"dbName"`
}

type ProcessConfig struct {
	Path string `yaml:"path" json:"path"`
}

type Config struct {
	Log        *LogConfig        `yaml:"log" json:"log"`
	Etcd       *EtcdConfig       `yaml:"etcd" json:"etcd"`
	Server     *GrpcServerConfig `yaml:"grpcServer" json:"grpcServer"`
	HttpServer *HttpServerConfig `yaml:"httpServer" json:"httpServer"`
	Database   *DatabaseConfig   `yaml:"database" json:"database"`
	Process    *ProcessConfig    `yaml:"process" json:"process"`
}

func (s *GrpcService) getConfig(configPath string) (*Config, error) {
	confData, err := ioutil.ReadFile(configPath)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(confData, &config); err != nil {
		logger.Error(err)
		return nil, err
	}

	data, _ := json.Marshal(&config)
	logger.Info("config:", string(data))
	return &config, nil
}
