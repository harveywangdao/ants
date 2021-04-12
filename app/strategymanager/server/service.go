package server

import (
	"net"
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
}

func NewStrategyManager(configPath string) (*StrategyManager, error) {
	s := &StrategyManager{}

	config, err := s.getConfig(configPath)
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	s.config = config

	dbParam := s.config.Database.Username + ":" + s.config.Database.Password + "@tcp(" + s.config.Database.Address + ")/" + s.config.Database.DbName + "?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(s.config.Database.DriverName, dbParam)
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	// defer db.Close()
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(time.Hour)
	s.db = db

	return s, nil
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
