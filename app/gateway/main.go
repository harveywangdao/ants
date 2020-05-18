package main

import (
	"github.com/harveywangdao/ants/app/gateway/service"
	"github.com/harveywangdao/ants/logger"
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
	logger.Info("Start Server")
	service.StartService()
	logger.Info("Stop Server")
}
