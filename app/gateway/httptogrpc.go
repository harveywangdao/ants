package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/harveywangdao/ants/logger"
	"github.com/harveywangdao/ants/register/discovery"

	//"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func httpToGrpc(dis *discovery.Discovery, svcName, grpcSvcName, method string, reqData []byte) ([]byte, error) {
	addrs, err := dis.QueryServiceIpPort(svcName)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(addrs) == 0 {
		logger.Error("can not find service", svcName)
		return nil, errors.New("can not find service " + svcName)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := r.Intn(len(addrs))

	logger.Info("client connect:", addrs[i])

	cc, err := grpc.Dial(addrs[i], grpc.WithInsecure())
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer cc.Close()

	reflectClient := grpcreflect.NewClient(context.Background(), reflectpb.NewServerReflectionClient(cc))
	svcs, err := reflectClient.ListServices()
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info("services:", svcs)

	fullSvcName := ""
	for _, svc := range svcs {
		if strings.Contains(svc, grpcSvcName) {
			fullSvcName = svc
			break
		}
	}

	if fullSvcName == "" {
		logger.Error("can not find grpc service", grpcSvcName)
		return nil, errors.New("can not find grpc service " + grpcSvcName)
	}

	svcDesc, err := reflectClient.ResolveService(fullSvcName)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info(svcDesc.String())

	for _, v := range svcDesc.GetMethods() {
		logger.Info(v.String())
		logger.Info(v.GetInputType().String())
		logger.Info(v.GetOutputType().String())
	}

	methodDesc := svcDesc.FindMethodByName(method)
	if methodDesc == nil {
		logger.Errorf("service %s does not include a method named %s", grpcSvcName, method)
		return nil, fmt.Errorf("service %s does not include a method named %s", grpcSvcName, method)
	}

	reqMsg := dynamic.NewMessage(methodDesc.GetInputType())
	reqMsg.UnmarshalJSON(reqData)

	stub := grpcdynamic.NewStub(cc)
	respMsg, err := stub.InvokeRpc(context.Background(), methodDesc, reqMsg)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info("respMsg:", respMsg)

	resp, ok := respMsg.(*dynamic.Message)
	if !ok {
		logger.Error("respMsg convert fail")
		return nil, errors.New("respMsg convert fail")
	}

	return resp.MarshalJSON()
}
