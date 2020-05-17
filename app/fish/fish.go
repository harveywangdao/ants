package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		log.Println("ping start")

		c.JSON(200, gin.H{
			"message": "wang pong",
		})
	})

	log.Println("ping service start")
	r.Run(":4596")
}
