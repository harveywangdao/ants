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
	Port string `yaml:"port" json:"port"`
}

type DatabaseConfig struct {
	Address    string `yaml:"address" json:"address"`
	Username   string `yaml:"username" json:"username"`
	Password   string `yaml:"password" json:"password"`
	DriverName string `yaml:"driverName" json:"driverName"`
	DbName     string `yaml:"dbName" json:"dbName"`
}

type ClientConfig struct {
	UserServiceName  string `yaml:"userServiceName" json:"userServiceName"`
	GoodsServiceName string `yaml:"goodsServiceName" json:"goodsServiceName"`
}

type RedisConfig struct {
	Address          string `yaml:"address" json:"address"`
	Password         string `yaml:"password" json:"password"`
	RedisLockTimeout int64  `yaml:"redisLockTimeout" json:"redisLockTimeout"`
}

type KafkaConfig struct {
	Addrs []string `yaml:"addrs" json:"addrs"`
}

type NsqConfig struct {
	Addrs []string `yaml:"addrs" json:"addrs"`
}

type Config struct {
	Log        *LogConfig        `yaml:"log" json:"log"`
	Etcd       *EtcdConfig       `yaml:"etcd" json:"etcd"`
	Server     *ServerConfig     `yaml:"server" json:"server"`
	HttpServer *HttpServerConfig `yaml:"httpServer" json:"httpServer"`
	Database   *DatabaseConfig   `yaml:"database" json:"database"`
	Client     *ClientConfig     `yaml:"client" json:"client"`
	Redis      *RedisConfig      `yaml:"redis" json:"redis"`
	Kafka      *KafkaConfig      `yaml:"kafka" json:"kafka"`
	Nsq        *NsqConfig        `yaml:"nsq" json:"nsq"`
}

func getConfig() (*Config, error) {
	confData, err := ioutil.ReadFile(configPath)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(confData, &config)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	data, _ := json.Marshal(&config)
	logger.Debug("config:", string(data))
	return &config, nil
}
