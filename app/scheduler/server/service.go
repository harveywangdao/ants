package server

import (
	"time"

	"github.com/harveywangdao/ants/app/scheduler/util/logger"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"go.etcd.io/etcd/clientv3"
)

type HttpService struct {
	config    *Config
	router    *gin.Engine
	db        *gorm.DB
	scheduler *Scheduler

	client *clientv3.Client
}

func NewHttpService(configPath string) (*HttpService, error) {
	s := &HttpService{}

	// 加载配置
	config, err := s.getConfig(configPath)
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	s.config = config

	s.router = gin.Default()
	s.setRouter()

	// 连接数据库
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

	// 连接etcd
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: s.config.Etcd.Endpoints,
	})
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	s.client = cli

	scheduler, err := NewScheduler(s.config.Etcd.Endpoints)
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	s.scheduler = scheduler

	return s, nil
}

func (s *HttpService) Start(port string) error {
	if err := s.router.Run(":" + port); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}
