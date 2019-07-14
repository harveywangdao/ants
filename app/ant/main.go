package main

import (
	"github.com/harveywangdao/ants/logger"
	"github.com/harveywangdao/ants/register"
	"github.com/harveywangdao/ants/register/discovery"
	"log"
	"time"
)

func init() {
	fileHandler := logger.NewFileHandler("ant.log")
	logger.SetHandlers(logger.Console, fileHandler)
	//logger.SetHandlers(logger.Console)
	//defer logger.Close()
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	logger.SetLevel(logger.INFO)
}

func main() {
	err := initConfig()
	if err != nil {
		logger.Error(err)
		return
	}

	reg, err := register.NewRegister(getConf().Etcd.Endpoints, getConf().Server.Port, getConf().Server.Name)
	if err != nil {
		logger.Error(err)
		return
	}
	reg.Start()

	dis, err := discovery.NewDiscovery(getConf().Etcd.Endpoints)
	if err != nil {
		logger.Error(err)
		return
	}

	go AntSaySvcStart(getConf().Server.Port)
	//go StartHttpServer(getConf().HttpServer.Port, getConf().Server.Name, "Hello", HelloHandler)
	time.Sleep(time.Second)
	go ClientStart(dis, getConf().Client.AntServiceName)
	//go StartHttpClient("192.168.1.7", getConf().HttpServer.Port, getConf().Server.Name, "Hello")

	for {
		time.Sleep(5 * time.Second)

		/*
			ipports, err := dis.QueryServiceIpPort(getConf().Server.Name)
			if err != nil {
				logger.Error(err)
				return
			}
			logger.Info("ipports:", ipports)
		*/

		dis.GetAllService()
	}

	logger.Info("Start Server")
}
