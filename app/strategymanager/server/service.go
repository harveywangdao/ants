package server

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/harveywangdao/ants/app/scheduler/util/logger"
	pb "github.com/harveywangdao/ants/app/strategymanager/protos/strategymanager"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc"
)

type StrategyManager struct {
	config *Config
	db     *gorm.DB

	processes map[string]*StrategyProcess
	mu        sync.RWMutex
}

func NewStrategyManager(configPath string) (*StrategyManager, error) {
	mgr := &StrategyManager{
		processes: make(map[string]*StrategyProcess),
	}

	config, err := mgr.getConfig(configPath)
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	mgr.config = config

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
