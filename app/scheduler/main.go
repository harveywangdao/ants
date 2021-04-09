package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	port := flag.String("port", "8080", "port")
	flag.Parse()

	// 设置路由
	router := gin.Default()

	// start http server
	router.Run(":" + *port)
}
