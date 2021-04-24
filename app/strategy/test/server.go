package main

import (
	"context"
	//"fmt"
	"log"
	"net"
	"os"
	"time"

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
	logger.Info("success recv start strategy")
	return &pb.StartStrategyResponse{}, nil
}

func (s *StrategyService) StopStrategy(ctx context.Context, in *pb.StopStrategyRequest) (*pb.StopStrategyResponse, error) {
	return &pb.StopStrategyResponse{}, nil
}

func (s *StrategyService) StrategyExec(ctx context.Context, in *pb.StrategyExecRequest) (*pb.StrategyExecResponse, error) {
	return &pb.StrategyExecResponse{}, nil
}

func do1() {
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

	go func() {
		time.Sleep(time.Second * 20)
		var p *int
		*p = 2
	}()

	s := grpc.NewServer()
	pb.RegisterStrategyServer(s, &StrategyService{})
	logger.Fatal(s.Serve(lis))
}

func do2() {
	lisFile := os.NewFile(uintptr(3), "/tmp/45615641465414")
	lis, err := net.FileListener(lisFile)
	if err != nil {
		logger.Fatal(err)
		return
	}

	logger.Info("lisFile.Fd() =", lisFile.Fd())
	lisFile.Close()
	//lis.Close()

	go func() {
		time.Sleep(time.Second * 20)
		var p *int
		*p = 2
	}()

	s := grpc.NewServer()
	pb.RegisterStrategyServer(s, &StrategyService{})
	logger.Fatal(s.Serve(lis))
}

func do7() {
	/*connFile := os.NewFile(uintptr(3), "/tmp/45615641465414")
	conn, err := net.FileConn(connFile)
	if err != nil {
		logger.Fatal(err)
		return
	}*/

	count := 0
	for {
		/*n, err := conn.Write([]byte(time.Now().String()))
		if err != nil {
			logger.Error(err)
		} else {
			logger.Infof("send msg %d bytes", n)
		}*/
		time.Sleep(time.Second * 2)
		count++

		if count == 5 {
			var p *int
			*p = 1
		}
	}
}

func main() {
	do7()
}
