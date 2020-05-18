package service

import (
	"encoding/json"
	"github.com/harveywangdao/ants/logger"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	configPath = "conf/app.yaml"
)

var config Config

type LogConfig struct {
	LogPath  string `yaml:"logPath" json:"logPath"`
	LogLevel string `yaml:"logLevel" json:"logLevel"`
}

type EtcdConfig struct {
	Endpoints []string `yaml:"endpoints" json:"endpoints"`
}

type ServerConfig struct {
	Name string `yaml:"name" json:"name"`
	Port string `yaml:"port" json:"port"`
}

type HttpServerConfig struct {
	Port      string `yaml:"port" json:"port"`
	PrefixUrl string `yaml:"prefixUrl" json:"prefixUrl"`
}

type Config struct {
	Log        *LogConfig        `yaml:"log" json:"log"`
	Etcd       *EtcdConfig       `yaml:"etcd" json:"etcd"`
	Server     *ServerConfig     `yaml:"server" json:"server"`
	HttpServer *HttpServerConfig `yaml:"httpServer" json:"httpServer"`
}

func initConfig() error {
	confData, err := ioutil.ReadFile(configPath)
	if err != nil {
		logger.Error(err)
		return err
	}

	err = yaml.Unmarshal(confData, &config)
	if err != nil {
		logger.Error(err)
		return err
	}

	data, _ := json.Marshal(&config)
	logger.Info("config:", string(data))
	return nil
}

func getConf() *Config {
	return &config
}
