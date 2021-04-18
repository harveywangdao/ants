package server

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/harveywangdao/ants/app/scheduler/model"
	"github.com/harveywangdao/ants/app/scheduler/util"
	"github.com/harveywangdao/ants/app/scheduler/util/logger"

	"github.com/gin-gonic/gin"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/clientv3util"
	"go.etcd.io/etcd/clientv3/concurrency"
	mvccpb "go.etcd.io/etcd/mvcc/mvccpb"
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
	// /strategy/task/$apikey/$strategy/$instrumentid --> {"uptime":111111111, "available":true, "meta_data"："xxxx"}
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

	strategyTaskPath := fmt.Sprintf("%s/%s", StrategyTaskPrefix, req.ApiKey)
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
		// TODO:是否在线
		info.Status = StrategyTaskStatusRunning // StrategyTaskStatusStop
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

	// TODO:策略服务更新参数

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	strategyTaskPath := fmt.Sprintf("%s/%s/%s/%s", StrategyTaskPrefix, req.ApiKey, req.StrategyName, req.Symbol)
	if _, err := s.client.Delete(ctx, strategyTaskPath); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}

	// TODO:如果在运行就停止策略
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

	if err := s.startStrategyTask(req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}
}

func (s *HttpService) startStrategyTask(req *StrategyData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	strategyTaskPath := fmt.Sprintf("%s/%s/%s/%s", StrategyTaskPrefix, req.ApiKey, req.StrategyName, req.Symbol)

	// 检查策略是否存在

	// 开始策略

	// etcd增加映射信息
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
	if err := s.stopStrategyTask(req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}
}

func (s *HttpService) stopStrategyTask(req *StrategyData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	strategyTaskPath := fmt.Sprintf("%s/%s/%s/%s", StrategyTaskPrefix, req.ApiKey, req.StrategyName, req.Symbol)

	// 检查策略是否在运行

	// 停止策略

	// etcd删除映射信息
}

const (
	EtcdApiKeyPrefix           = "/scheduler/apikey/"
	EtcdApiKeyToStrategyPrefix = "/scheduler/apikey_strategy/"
	EtcdStrategyRegisterPrefix = "/service/strategy/"
)

type StrategyParam struct {
	ApiKey    string `json:"apiKey"`
	SecretKey string `json:"secretKey"`
	Strategy  string `json:"strategyName"`
}

type StrategyNode struct {
	Node     string `json:"node"`
	Strategy string `json:"strategy"`
}

type NodeInfo struct {
	ApiKeys map[string]int `json"apiKeys"`
}

func (s *Scheduler) StartStrategyTask(apiKey, secretKey, strategy string) {
	s.startCh <- &StrategyParam{
		ApiKey:    apiKey,
		SecretKey: secretKey,
		Strategy:  strategy,
	}
}

func (s *Scheduler) StopStrategyTask(apiKey string) {
	s.stopCh <- &StrategyParam{
		ApiKey: apiKey,
	}
}

func (s *Scheduler) apiKeyTask() {
	for {
		select {
		case <-s.closed:
			logger.Info("scheduler exit")
			return

		case st := <-s.startCh:
			// 向etcd注册信息/scheduler/apikey/$apikey --> {"strategyName":"grid","param":0.1}
			data, err := json.Marshal(st)
			if err != nil {
				logger.Error(err)
				continue
			}
			key := EtcdApiKeyPrefix + st.ApiKey
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			resp, err := s.client.Put(ctx, key, string(data))
			if err != nil {
				logger.Errorf("put etcd fail, key: %s, err: %v", key, err)
			} else {
				logger.Infof("put etcd success, key: %s", key)
				logger.Info(resp)
			}
			cancel()

		case st := <-s.stopCh:
			key := EtcdApiKeyPrefix + st.ApiKey
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			resp, err := s.client.Delete(ctx, key)
			if err != nil {
				logger.Errorf("delete etcd fail, key: %s, err: %v", key, err)
			} else {
				logger.Infof("delete etcd success, key: %s", key)
				logger.Info(resp)
			}
			cancel()
		}
	}
}

func (s *Scheduler) watchStrategyRegisterTask(stopMasterCh chan int) {
	for {
		select {
		case <-s.closed:
			logger.Info("watchStrategyNode exit")
			return
		case <-stopMasterCh:
			logger.Info("watchStrategyNode exit")
			return
		default:
		}

		s.watchStrategyRegister(stopMasterCh)
	}
}

func (s *Scheduler) watchStrategyRegister(stopMasterCh chan int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// /service/strategy/strategyName/10.22.33.55:32154 --> {"uptime":111111111, "available":true, "strategyName":"grid"}
	rch := s.client.Watch(ctx, EtcdStrategyRegisterPrefix, clientv3.WithPrefix())
	for {
		select {
		case wresp, ok := <-rch:
			if !ok {
				logger.Errorf("watch %s exit", EtcdStrategyRegisterPrefix)
				return
			}

			for _, ev := range wresp.Events {
				logger.Infof("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				strategyAndNodeStr := strings.TrimPrefix(string(ev.Kv.Key), EtcdStrategyRegisterPrefix)
				strategyAndNode := strings.Split(strategyAndNodeStr, "/")
				if len(strategyAndNode) != 2 {
					logger.Errorf("%s path error", string(ev.Kv.Key))
					continue
				}
				strategy := strategyAndNode[0]
				nodeAddr := strategyAndNode[1]

				switch ev.Type {
				case mvccpb.PUT:
					s.strategyNodeOnline(strategy, nodeAddr)
				case mvccpb.DELETE:
					s.strategyNodeOffline(strategy, nodeAddr)
				}
			}

		case <-s.closed:
			return
		case <-stopMasterCh:
			return
		}
	}
}

func (s *Scheduler) strategyNodeOnline(strategy, nodeAddr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 增加 /scheduler/strategy/strategyName/10.22.33.55:32154 --> {}
	data, err := json.Marshal(&NodeInfo{})
	if err != nil {
		logger.Error(err)
		return err
	}
	strategyNodePath := EtcdStrategyNodePrefix + strategy + "/" + nodeAddr
	if _, err = s.client.Put(ctx, strategyNodePath, string(data)); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (s *Scheduler) strategyNodeOffline(strategy, nodeAddr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取此节点上有哪些策略 /scheduler/strategy/strategyName/10.22.33.55:32154 --> {"apiKeys":["apikey1","apikey2"]}
	strategyNodePath := EtcdStrategyNodePrefix + strategy + "/" + nodeAddr
	getResp, err := s.client.Get(ctx, strategyNodePath)
	if err != nil {
		logger.Error(err)
		return err
	}
	if len(getResp.Kvs) == 0 {
		return fmt.Errorf("%s not existed", strategyNodePath)
	}
	ni := &NodeInfo{}
	if err := json.Unmarshal(getResp.Kvs[0].Value, ni); err != nil {
		logger.Error(err)
		return err
	}

	// 删除 /scheduler/strategy/strategyName/10.22.33.55:32154 --> {"apiKeys":["apikey1","apikey2"]}
	if _, err := s.client.Delete(ctx, strategyNodePath); err != nil {
		logger.Error(err)
		return err
	}

	// 重新调度策略
	for apiKey, _ := range ni.ApiKeys {
		// 删除 /scheduler/apikey_strategy/$apikey --> {"node":"10.22.33.55:32154"}
		apiKeyToStrategyPath := EtcdApiKeyToStrategyPrefix + apiKey
		if _, err := s.client.Delete(ctx, apiKeyToStrategyPath); err != nil {
			logger.Error(err)
			continue
		}

		// 获取 /scheduler/apikey/$apikey --> {"strategyName":"grid","param":0.1}
		getResp, err := s.client.Get(ctx, EtcdApiKeyPrefix+apiKey)
		if err != nil {
			logger.Error(err)
			continue
		}
		if len(getResp.Kvs) == 0 {
			continue
		}

		param := &StrategyParam{}
		if err := json.Unmarshal(getResp.Kvs[0].Value, param); err != nil {
			logger.Error(err)
			continue
		}

		s.startStrategy(apiKey, param)
	}

	return nil
}

func (s *Scheduler) watchApiKeyTask(stopMasterCh chan int) {
	for {
		select {
		case <-s.closed:
			logger.Info("watchApiKeyTask exit")
			return
		case <-stopMasterCh:
			logger.Info("watchApiKeyTask exit")
			return
		default:
		}

		s.watchApiKey(stopMasterCh)
	}
}

func (s *Scheduler) watchApiKey(stopMasterCh chan int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// /scheduler/apikey/$apikey --> {"strategyName":"grid","param":0.1}
	rch := s.client.Watch(ctx, EtcdApiKeyPrefix, clientv3.WithPrefix())
	for {
		select {
		case wresp, ok := <-rch:
			if !ok {
				logger.Errorf("watch %s exit", EtcdApiKeyPrefix)
				return
			}

			for _, ev := range wresp.Events {
				logger.Infof("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				apiKey := strings.TrimPrefix(string(ev.Kv.Key), EtcdApiKeyPrefix)

				switch ev.Type {
				case mvccpb.PUT:
					param := &StrategyParam{}
					if err := json.Unmarshal(ev.Kv.Value, param); err != nil {
						logger.Error(err)
						continue
					}
					s.startOrUpdateStrategy(apiKey, param)
				case mvccpb.DELETE:
					s.stopStrategy(apiKey)
				}
			}

		case <-s.closed:
			return
		case <-stopMasterCh:
			return
		}
	}
}

func (s *Scheduler) startOrUpdateStrategy(apiKey string, param *StrategyParam) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// /scheduler/apikey_strategy/$apikey --> {"node":"10.22.33.55:32154"}
	getResp, err := s.client.Get(ctx, EtcdApiKeyToStrategyPrefix+apiKey)
	if err != nil {
		logger.Error(err)
		return err
	}

	if len(getResp.Kvs) == 0 {
		// 增加
		return s.startStrategy(apiKey, param)
	}

	// 更新
	sn := &StrategyNode{}
	if err := json.Unmarshal(getResp.Kvs[0].Value, sn); err != nil {
		logger.Error(err)
		return err
	}
	if param.Strategy == sn.Strategy {
		// 策略和参数未变,不用更新
		return nil
	}

	// 停止旧策略
	if err := s.stopStrategy(apiKey); err != nil {
		logger.Error(err)
		return err
	}
	// 启动新策略
	if err := s.startStrategy(apiKey, param); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (s *Scheduler) startStrategy(apiKey string, param *StrategyParam) error {
	addr := s.getAppropriateNode(param.Strategy)
	if addr == "" {
		return fmt.Errorf("can not find node, strategy: %s", param.Strategy)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 增加apikey到策略节点的映射关系
	data, err := json.Marshal(&StrategyNode{
		Node:     addr,
		Strategy: param.Strategy,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	// /scheduler/apikey_strategy/$apikey --> {"node":"10.22.33.55:32154"}
	if _, err = s.client.Put(ctx, EtcdApiKeyToStrategyPrefix+apiKey, string(data)); err != nil {
		logger.Error(err)
		return err
	}

	// 更新注册节点上的apikey的信息
	// /scheduler/strategy/strategyName/10.22.33.55:32154 --> {"apiKeys":["apikey1","apikey2"]}
	strategyNodePath := EtcdStrategyNodePrefix + param.Strategy + "/" + addr
	getResp, err := s.client.Get(ctx, strategyNodePath)
	if err != nil {
		logger.Error(err)
		return err
	}
	if len(getResp.Kvs) == 0 {
		return fmt.Errorf("%s not existed", strategyNodePath)
	}

	ni := &NodeInfo{}
	if err := json.Unmarshal(getResp.Kvs[0].Value, ni); err != nil {
		logger.Error(err)
		return err
	}
	if ni.ApiKeys == nil {
		ni.ApiKeys = make(map[string]int)
	}
	ni.ApiKeys[apiKey] = 1

	data, err = json.Marshal(ni)
	if err != nil {
		logger.Error(err)
		return err
	}
	if _, err = s.client.Put(ctx, strategyNodePath, string(data)); err != nil {
		logger.Error(err)
		return err
	}

	// TODO: connect to strategy node and start

	return nil
}

func (s *Scheduler) stopStrategy(apiKey string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 删除apikey到策略节点的映射关系
	// /scheduler/apikey_strategy/$apikey --> {"node":"10.22.33.55:32154"}
	apiKeyToStrategyPath := EtcdApiKeyToStrategyPrefix + apiKey
	getResp, err := s.client.Get(ctx, apiKeyToStrategyPath)
	if err != nil {
		logger.Error(err)
		return err
	}
	if len(getResp.Kvs) == 0 {
		return fmt.Errorf("%s not existed", apiKeyToStrategyPath)
	}

	sn := &StrategyNode{}
	if err := json.Unmarshal(getResp.Kvs[0].Value, sn); err != nil {
		logger.Error(err)
		return err
	}
	if _, err := s.client.Delete(ctx, apiKeyToStrategyPath); err != nil {
		logger.Error(err)
		return err
	}

	// 更新注册节点上的apikey的信息
	// /scheduler/strategy/strategyName/10.22.33.55:32154 --> {"apiKeys":["apikey1","apikey2"]}
	strategyNodePath := EtcdStrategyNodePrefix + sn.Strategy + "/" + sn.Node
	getResp, err = s.client.Get(ctx, strategyNodePath)
	if err != nil {
		logger.Error(err)
		return err
	}
	if len(getResp.Kvs) == 0 {
		return fmt.Errorf("%s not existed", strategyNodePath)
	}

	ni := &NodeInfo{}
	if err := json.Unmarshal(getResp.Kvs[0].Value, ni); err != nil {
		logger.Error(err)
		return err
	}
	delete(ni.ApiKeys, apiKey)

	data, err := json.Marshal(ni)
	if err != nil {
		logger.Error(err)
		return err
	}
	if _, err = s.client.Put(ctx, strategyNodePath, string(data)); err != nil {
		logger.Error(err)
		return err
	}

	// TODO: stop strategy
	return nil
}
