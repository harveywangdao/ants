package main

import (
	"context"
	"github.com/harveywangdao/ants/logger"
	antpb "github.com/harveywangdao/ants/rpc/ant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

type AntSaySvc struct{}

func (s *AntSaySvc) Hello(ctx context.Context, in *antpb.HelloRequest) (*antpb.HelloResponse, error) {
	logger.Info("service recv:", in.Name)
	return &antpb.HelloResponse{Msg: "Hello " + in.Name + " at " + time.Now().Format("2006-01-02 15:04:05")}, nil
}

func AntSaySvcStart(port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error(err)
		return
	}

	s := grpc.NewServer()
	antpb.RegisterAntSayServer(s, &AntSaySvc{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		logger.Error("service halt! error:", err)
		return
	}
}
