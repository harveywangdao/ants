package service

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/harveywangdao/ants/logger"
	"github.com/harveywangdao/ants/register"
	"github.com/harveywangdao/ants/register/discovery"
	proto "github.com/harveywangdao/ants/rpc/goods"
	"github.com/harveywangdao/ants/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/jinzhu/gorm"
)

type Service struct {
	Config    *Config
	discovery *discovery.Discovery
	db        *gorm.DB
}

var (
	App = &Service{}
)

func initService() error {
	// config
	config, err := getConfig()
	if err != nil {
		logger.Error(err)
		return err
	}
	App.Config = config

	// set logger
	dir := filepath.Dir(App.Config.Log.LogPath)
	if dir != "" && !util.IsDir(dir) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	fileHandler := logger.NewFileHandler(App.Config.Log.LogPath)
	logger.SetHandlers(logger.Console, fileHandler)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	logger.SetLoggerLevel(App.Config.Log.LogLevel)

	// discovery
	dis, err := discovery.NewDiscovery(App.Config.Etcd.Endpoints)
	if err != nil {
		logger.Error(err)
		return err
	}
	App.discovery = dis

	dbConfig := App.Config.Database
	dbParam := dbConfig.Username + ":" + dbConfig.Password + "@tcp(" + dbConfig.Address + ")/" + dbConfig.DbName + "?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(dbConfig.DriverName, dbParam)
	if err != nil {
		logger.Error(err)
		return err
	}
	// defer db.Close()
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(time.Hour)

	App.db = db

	return nil
}

func StartHttpService() error {
	httpService := &HttpService{
		ServiceApp: App,
	}

	httpServer := &HttpServer{
		ServiceName: App.Config.Server.Name,
		Port:        App.Config.HttpServer.Port,
	}

	go httpServer.StartHttpServer(httpService)

	return nil
}

func StartService() error {
	if err := initService(); err != nil {
		logger.Error(err)
		return err
	}

	reg, err := register.NewRegister(App.Config.Etcd.Endpoints, App.Config.Server.Port, App.Config.Server.Name)
	if err != nil {
		logger.Error(err)
		return err
	}
	reg.Start()

	if err := StartHttpService(); err != nil {
		logger.Error(err)
		return err
	}

	lis, err := net.Listen("tcp", ":"+App.Config.Server.Port)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("rpc server:", lis.Addr())

	s := grpc.NewServer()
	proto.RegisterGoodsServiceServer(s, App)
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		logger.Error("service halt! error:", err)
		return err
	}

	return nil
}
