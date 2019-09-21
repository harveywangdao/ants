package main

import (
	"log"

	_ "github.com/go-sql-driver/mysql"

	"github.com/harveywangdao/ants/app/order/service"
	"github.com/harveywangdao/ants/logger"
)

func init() {
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
