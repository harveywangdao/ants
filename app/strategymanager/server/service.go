package server

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/harveywangdao/ants/app/scheduler/util/logger"
	pb "github.com/harveywangdao/ants/app/strategymanager/protos/strategymanager"
	"google.golang.org/grpc"
)

type StrategyManager struct {
	config *Config

	processes  map[string]*StrategyProcess
	mu         sync.RWMutex
	procExitCh chan string

	once sync.Once
}

func NewStrategyManager(configPath string) (*StrategyManager, error) {
	mgr := &StrategyManager{
		processes:  make(map[string]*StrategyProcess),
		procExitCh: make(chan string),
	}

	config, err := mgr.getConfig(configPath)
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	mgr.config = config

	go mgr.monitor()

	return mgr, nil
}

func (s *StrategyManager) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("rpc server:", lis.Addr())

	srv := grpc.NewServer()
	pb.RegisterStrategyManagerServer(srv, s)
	reflection.Register(srv)

	if err := srv.Serve(lis); err != nil {
		logger.Error("service halt! error:", err)
		return err
	}

	return nil
}

func (s *StrategyManager) monitor() {
	for {
		select {
		case uniqueId := <-s.procExitCh:
			s.delProc(uniqueId)
		}
	}
}
