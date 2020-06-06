package service

import (
	"github.com/harveywangdao/ants/logger"
)

func StartService() error {
	err := initConfig()
	if err != nil {
		logger.Error(err)
		return err
	}

	StartHttpServer()
	return nil
}
