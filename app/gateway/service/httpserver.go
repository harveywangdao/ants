package service

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
	logger.Info("body:", string(body))

	dis, err := discovery.NewDiscovery(getConf().Etcd.Endpoints)
	if err != nil {
		logger.Error(err)
		return
	}

	if r.Form.Get(svc) == "" || r.Form.Get(grpcsvc) == "" || r.Form.Get(method) == "" {
		logger.Error("param error")
		return
	}

	resp, err := httpToGrpc(dis, r.Form.Get(svc), r.Form.Get(grpcsvc), r.Form.Get(method), body)
	if err != nil {
		logger.Error(err)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(resp)
}

func listenHttp(port, prefixUrl string) {
	mux := http.NewServeMux()
	mux.Handle(prefixUrl, &gatewayHandler{})
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("http pong ants"))
	})

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

func listenHttps(port, prefixUrl string) {
	mux := http.NewServeMux()
	mux.Handle(prefixUrl, &gatewayHandler{})
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("https pong ants"))
	})

	server := &http.Server{
		Addr:         ":" + port,
		WriteTimeout: time.Second * 3, //设置3秒的写超时
		Handler:      mux,
	}

	if err := server.ListenAndServeTLS("ca/server.crt", "ca/server.key"); err != nil {
		logger.Error(err)
		return
	}
}

func StartHttpServer() {
	go listenHttp(getConf().HttpServer.Port, getConf().HttpServer.PrefixUrl)
	listenHttps(getConf().HttpServer.HttpsPort, getConf().HttpServer.PrefixUrl)
}
