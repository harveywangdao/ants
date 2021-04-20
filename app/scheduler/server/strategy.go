package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/harveywangdao/ants/app/scheduler/util/logger"

	"github.com/gin-gonic/gin"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/clientv3util"
)

const (
	StrategyTaskPrefix = "/strategy/task"

	StrategyTaskStatusRunning = "running"
	StrategyTaskStatusStop    = "stop"
)

type StrategyData struct {
	StrategyName  string `json:"strategy_name"`
	Exchange      string `json:"exchange"`
	ApiKey        string `json:"api_key"`
	SecretKey     string `json:"secret_key"`
	Passphrase    string `json:"passphrase"`
	Symbol        string `json:"symbol"`
	Commission    string `json:"commission"`
	InitialRights string `json:"initial_rights"`
	Params        string `json:"params"`
	UserId        string `json:"user_id"`
}

type StrategyTaskInfo struct {
	MetaData  *StrategyData `json:"meta_data"`
	Uptime    int64         `json:"uptime"`
	Available bool          `json:"available"`
	Status    string        `json:"status"` // running stop
}

func (s *HttpService) AddStrategyTask(c *gin.Context) {
	req := &StrategyData{}
	if err := c.BindJSON(req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.ApiKey == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	if err := s.addStrategyTask(c, req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *HttpService) addStrategyTask(c *gin.Context, req *StrategyData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data, err := json.Marshal(&StrategyTaskInfo{
		MetaData:  req,
		Uptime:    time.Now().Unix(),
		Available: true,
	})
	if err != nil {
		logger.Error(err)
		return err
	}

	// /strategy/task/$apikey/$strategy/$instrumentid --> {"uptime":111111111, "available":true, "meta_data"："xxxx"}
	strategyTaskPath := fmt.Sprintf("%s/%s/%s/%s", StrategyTaskPrefix, req.ApiKey, req.StrategyName, req.Symbol)
	kvc := clientv3.NewKV(s.client)
	txnResp, err := kvc.Txn(ctx).
		If(clientv3util.KeyMissing(strategyTaskPath)).
		Then(clientv3.OpPut(strategyTaskPath, string(data))).
		Commit()
	if err != nil {
		logger.Error(err)
		return err
	}
	if !txnResp.Succeeded {
		return fmt.Errorf("%s already existed", strategyTaskPath)
	}

	return nil
}

func (s *HttpService) QueryStrategyTasks(c *gin.Context) {
	apikey := c.Query("apikey")
	if apikey == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// /strategy/task/$apikey/$strategy/$instrumentid --> {"uptime":111111111, "available":true, "meta_data"："xxxx"}
	strategyTaskPath := fmt.Sprintf("%s/%s", StrategyTaskPrefix, apikey)
	getResp, err := s.client.Get(ctx, strategyTaskPath, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}

	var tasks []*StrategyTaskInfo
	for _, kv := range getResp.Kvs {
		info := &StrategyTaskInfo{}
		if err := json.Unmarshal(kv.Value, info); err != nil {
			logger.Error(err)
			AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
			return
		}

		parts := strings.Split(strings.TrimPrefix(string(kv.Key), StrategyTaskPrefix+"/"), "/")
		if len(parts) != 3 {
			logger.Error("error key:", string(kv.Key))
			continue
		}

		info.Status = StrategyTaskStatusStop
		running, err := s.scheduler.IsTaskRunning(&StrategyData{
			ApiKey:       parts[0],
			StrategyName: parts[1],
			Symbol:       parts[2],
		})
		if err != nil {
			logger.Error(err)
			AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
			return
		}
		if running {
			info.Status = StrategyTaskStatusRunning
		}

		tasks = append(tasks, info)
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
	})
}

func (s *HttpService) UpdateStrategyTask(c *gin.Context) {
	req := &StrategyData{}
	if err := c.BindJSON(req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.ApiKey == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	if err := s.updateStrategyTask(c, req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *HttpService) updateStrategyTask(c *gin.Context, req *StrategyData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	running, err := s.scheduler.IsTaskRunning(req)
	if err != nil {
		logger.Error(err)
		return err
	}
	if running {
		if err := s.scheduler.UpdateOneStrategyTask(req); err != nil {
			logger.Error(err)
			return err
		}
	}

	data, err := json.Marshal(&StrategyTaskInfo{
		MetaData:  req,
		Uptime:    time.Now().Unix(),
		Available: true,
	})
	if err != nil {
		logger.Error(err)
		return err
	}

	strategyTaskPath := fmt.Sprintf("%s/%s/%s/%s", StrategyTaskPrefix, req.ApiKey, req.StrategyName, req.Symbol)
	kvc := clientv3.NewKV(s.client)
	txnResp, err := kvc.Txn(ctx).
		If(clientv3util.KeyExists(strategyTaskPath)).
		Then(clientv3.OpPut(strategyTaskPath, string(data))).
		Commit()
	if err != nil {
		logger.Error(err)
		return err
	}
	if !txnResp.Succeeded {
		return fmt.Errorf("%s already not existed", strategyTaskPath)
	}

	return nil
}

func (s *HttpService) DelStrategyTask(c *gin.Context) {
	req := &StrategyData{}
	if err := c.BindJSON(req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.ApiKey == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	running, err := s.scheduler.IsTaskRunning(req)
	if err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
	if running {
		if err := s.scheduler.StopOneStrategyTask(req); err != nil {
			logger.Error(err)
			AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	strategyTaskPath := fmt.Sprintf("%s/%s/%s/%s", StrategyTaskPrefix, req.ApiKey, req.StrategyName, req.Symbol)
	if _, err := s.client.Delete(ctx, strategyTaskPath); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *HttpService) StartStrategyTask(c *gin.Context) {
	req := &StrategyData{}
	if err := c.BindJSON(req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.ApiKey == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}
	if err := s.scheduler.StartOneStrategyTask(req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *HttpService) StopStrategyTask(c *gin.Context) {
	req := &StrategyData{}
	if err := c.BindJSON(req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.ApiKey == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}
	if err := s.scheduler.StopOneStrategyTask(req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}
