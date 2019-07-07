package main

import (
	"github.com/harveywangdao/ants/logger"
	"github.com/harveywangdao/ants/register/discovery"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	svc     = "svc"
	grpcsvc = "grpcsvc"
	method  = "method"
)

type gatewayHandler struct{}

func (s *gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	logger.Info(r.Form)
	logger.Info(r.URL.Path)
	logger.Info(r.Method)

	for k, v := range r.Form {
		logger.Info(k, "=>", v, strings.Join(v, "-"))
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error(err)
		return
	}
	//r.Body.Close()
	logger.Info("body:", string(body))

	dis, err := discovery.NewDiscovery(getConf().Etcd.Endpoints)
	if err != nil {
		logger.Error(err)
		return
	}

	if r.Form.Get(svc) == "" || r.Form.Get(grpcsvc) == "" || r.Form.Get(method) == "" {
		logger.Error("param lack")
		return
	}

	resp, err := httpToGrpc(dis, r.Form.Get(svc), r.Form.Get(grpcsvc), r.Form.Get(method), body)
	if err != nil {
		logger.Error(err)
		return
	}

	w.Write(resp)
}

func StartHttpServer(port, prefixUrl string) {
	mux := http.NewServeMux()
	mux.Handle(prefixUrl, &gatewayHandler{})

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
