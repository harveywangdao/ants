package main

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		client := &http.Client{}
		//url := "http://10.100.244.221:8080/ping"

		url := "http://app01-service.default.svc.cluster.local:8080/ping"

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Println(err)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}

		c.JSON(200, gin.H{
			"message": "shabi app02 " + string(body),
		})
	})
	r.Run(":8081") // listen and serve on 0.0.0.0:8081
}
