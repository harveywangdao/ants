package main

import (
	"github.com/harveywangdao/ants/logger"
	"net/http"
	"time"
)

type serviceHandler struct{}

func (s *serviceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("this is version 3"))
}

func StartHttpServer(port, serviceName, funcName string, handler func(http.ResponseWriter, *http.Request)) {
	mux := http.NewServeMux()
	mux.HandleFunc("/"+serviceName+"/"+funcName, handler)
	mux.Handle("/"+serviceName, &serviceHandler{})

	server := &http.Server{
		Addr:         ":" + port,
		WriteTimeout: time.Second * 3, //设置3秒的写超时
		Handler:      mux,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error(err)
		return
	}
}
