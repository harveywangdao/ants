package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/harveywangdao/ants/app/scheduler/util/logger"
	//"github.com/harveywangdao/ants/app/strategy/service"
	pb "github.com/harveywangdao/ants/app/strategymanager/protos/strategy"
	"google.golang.org/grpc"
)

func init() {
	logger.SetHandlers(logger.Console)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	logger.SetLevel(logger.INFO)
}

type StrategyService struct {
}

func (s *StrategyService) StartStrategy(ctx context.Context, in *pb.StartStrategyRequest) (*pb.StartStrategyResponse, error) {
	return &pb.StartStrategyResponse{}, nil
}

func (s *StrategyService) StopStrategy(ctx context.Context, in *pb.StopStrategyRequest) (*pb.StopStrategyResponse, error) {
	return &pb.StopStrategyResponse{}, nil
}

func (s *StrategyService) StrategyExec(ctx context.Context, in *pb.StrategyExecRequest) (*pb.StrategyExecResponse, error) {
	return &pb.StrategyExecResponse{}, nil
}

func main() {
	if len(os.Args) != 2 {
		logger.Fatal("args len must be 2, args:", os.Args)
		return
	}
	unixFile := os.Args[1]
	if err := os.RemoveAll(unixFile); err != nil {
		logger.Fatal(err)
		return
	}

	addr, err := net.ResolveUnixAddr("unix", unixFile)
	if err != nil {
		logger.Fatal(err)
		return
	}
	lis, err := net.ListenUnix("unix", addr)
	if err != nil {
		logger.Fatal(err)
		return
	}
	s := grpc.NewServer()
	pb.RegisterStrategyServer(s, &StrategyService{})
	logger.Fatal(s.Serve(lis))
}
