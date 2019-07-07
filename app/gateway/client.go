package main

import (
	"context"
	"github.com/harveywangdao/ants/logger"
	"github.com/harveywangdao/ants/register/discovery"
	antpb "github.com/harveywangdao/ants/rpc/ant"
	"google.golang.org/grpc"
	"math/rand"
	"time"
)

func ClientStart(dis *discovery.Discovery, svcName string) {
	addrs, err := dis.QueryServiceIpPort(svcName)
	if err != nil {
		logger.Error(err)
		return
	}

	if len(addrs) == 0 {
		logger.Error("can not find service", svcName)
		return
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := r.Intn(len(addrs))

	logger.Info("client connect:", addrs[i])

	conn, err := grpc.Dial(addrs[i], grpc.WithInsecure())
	if err != nil {
		logger.Error(err)
		return
	}
	defer conn.Close()

	c := antpb.NewAntSayClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	helloResp, err := c.Hello(ctx, &antpb.HelloRequest{Name: "xiaoming"})
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Info("client recv:", helloResp.Msg)
}
