package server

import (
	"net"
	"sync"

	"github.com/harveywangdao/ants/app/scheduler/util/logger"
	pb "github.com/harveywangdao/ants/app/strategymanager/protos/strategymanager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type StrategyManager struct {
	Config *Config

	processes map[string]*StrategyProcess
	mu        sync.RWMutex

	closeCh chan bool
}

func NewStrategyManager(configPath string) (*StrategyManager, error) {
	mgr := &StrategyManager{
		processes: make(map[string]*StrategyProcess),
		closeCh:   make(chan bool),
	}

	config, err := mgr.getConfig(configPath)
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	mgr.Config = config

	return mgr, nil
}

func (s *StrategyManager) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("rpc server:", lis.Addr())

	go s.registerTask(lis.Addr().(*net.TCPAddr).Port)

	srv := grpc.NewServer()
	pb.RegisterStrategyManagerServer(srv, s)
	reflection.Register(srv)

	if err := srv.Serve(lis); err != nil {
		logger.Error("service halt! error:", err)
		return err
	}

	return nil
}
