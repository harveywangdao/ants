package main

import (
	"ants/logger"
	antpb "ants/rpc/ant"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	logger.Info(r.Form)
	logger.Info(r.URL.Path)

	for k, v := range r.Form {
		logger.Info(k, "=>", v, strings.Join(v, "-"))
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("body:", string(body))

	var in antpb.HelloRequest
	err = json.Unmarshal(body, &in)
	if err != nil {
		logger.Error(err)
		return
	}

	svc := &AntSaySvc{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := svc.Hello(ctx, &in)
	if err != nil {
		logger.Error(err)
		return
	}

	data, _ := json.Marshal(out)

	fmt.Fprint(w, string(data))
}
