package main

import (
	"github.com/harveywangdao/ants/logger"
	"github.com/harveywangdao/ants/register"
	"log"
)

func init() {
	//fileHandler := logger.NewFileHandler("test.log")
	//logger.SetHandlers(logger.Console, fileHandler)
	logger.SetHandlers(logger.Console)
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

	go StartHttpServer(getConf().HttpServer.Port, getConf().HttpServer.PrefixUrl)

	select {}
	logger.Info("Server exit")
}
