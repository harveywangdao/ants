package server

import (
	"scheduler/util/logger"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type HttpService struct {
	config *Config
	router *gin.Engine
	db     *gorm.DB
}

func NewHttpService(configPath string) (*HttpService, error) {
	s := &HttpService{}

	config, err := s.getConfig(configPath)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	s.config = config

	s.router = gin.Default()
	s.setRouter()

	dbParam := s.config.Username + ":" + s.config.Password + "@tcp(" + s.config.Address + ")/" + s.config.DbName + "?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(s.config.DriverName, dbParam)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	// defer db.Close()
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(time.Hour)
	s.db = db

	return s, nil
}

func (s *HttpService) Start(port string) error {
	if err := s.router.Run(":" + port); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}
