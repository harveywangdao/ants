package main

import (
	"flag"
	"log"

	"github.com/harveywangdao/ants/app/scheduler/server"
	"github.com/harveywangdao/ants/app/scheduler/util/logger"
)

func init() {
	logger.SetHandlers(logger.Console)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	logger.SetLevel(logger.INFO)
}

func main() {
	port := flag.String("port", "8080", "port")
	configPath := flag.String("configPath", "conf/app.yaml", "port")
	flag.Parse()

	srv, err := server.NewHttpService(*configPath)
	if err != nil {
		logger.Fatal(err)
		return
	}
	logger.Fatal(srv.Start(*port))
}
