package main

import (
	"ants/logger"
	antpb "ants/rpc/ant"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func StartHttpClient(ip, port, serviceName, funcName string) {
	client := &http.Client{}

	helloRequest := &antpb.HelloRequest{
		Name: "xiaohong",
	}

	data, _ := json.Marshal(helloRequest)
	url := fmt.Sprintf("http://%s:%s/%s/%s", ip, port, serviceName, funcName)

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		logger.Error(err)
		return
	}

	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Set("Cookie", "name=anny")
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Info(string(body))
}
