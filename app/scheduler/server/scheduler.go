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
	"go.etcd.io/etcd/clientv3/concurrency"
	mvccpb "go.etcd.io/etcd/mvcc/mvccpb"
)

type StrategyReq struct {
	ApiKey string `json:"apiKey"`
}

func (s *HttpService) StartApikeyStrategy(c *gin.Context) {
	req := &StrategyReq{}
	if err := c.BindJSON(req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.ApiKey == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	var ak model.ApiKeyModel
	if err := s.db.Where("api_key = ?", req.ApiKey).First(&ak).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	s.scheduler.StartOneApiKeyStrategy(ak.ApiKey, ak.SecretKey, ak.Strategy)
}

func (s *HttpService) StopApikeyStrategy(c *gin.Context) {
	req := &StrategyReq{}
	if err := c.BindJSON(req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.ApiKey == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	s.scheduler.StopOneApiKeyStrategy(req.ApiKey)
}

func (s *HttpService) PauseStrategy(c *gin.Context) {

}

func (s *HttpService) ResumeStrategy(c *gin.Context) {

}

func (s *HttpService) MigrateStrategy(c *gin.Context) {

}

const (
	EtcdApiKeyPrefix           = "/scheduler/apikey/"
	EtcdApiKeyToStrategyPrefix = "/scheduler/apikey_strategy/"
	EtcdStrategyNodePrefix     = "/scheduler/strategy/"
	EtcdStrategyRegisterPrefix = "/service/strategy/"

	EtcdStrategyElectionKey = "/scheduler/master"
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

type Scheduler struct {
	startCh chan *StrategyParam
	stopCh  chan *StrategyParam

	client *clientv3.Client

	closed chan bool
	once   sync.Once

	nodes sync.Map
}

func NewScheduler(etcdAddrs []string) (*Scheduler, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: etcdAddrs,
	})
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	s := &Scheduler{
		client:  cli,
		startCh: make(chan *StrategyParam),
		stopCh:  make(chan *StrategyParam),
		closed:  make(chan bool),
	}
	go s.apiKeyTask()
	go s.masterRun()
	return s, nil
}

func (s *Scheduler) StartOneApiKeyStrategy(apiKey, secretKey, strategy string) {
	s.startCh <- &StrategyParam{
		ApiKey:    apiKey,
		SecretKey: secretKey,
		Strategy:  strategy,
	}
}

func (s *Scheduler) StopOneApiKeyStrategy(apiKey string) {
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

// 一主多从结构,使用etcd的选举
func (s *Scheduler) masterRun() {
	session, err := concurrency.NewSession(s.client)
	if err != nil {
		logger.Fatal(err)
		return
	}
	election := concurrency.NewElection(session, EtcdStrategyElectionKey)
	nodeName := util.GetUUID()
	logger.Info("current node name:", nodeName)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = election.Campaign(ctx, nodeName)
		cancel()
		if err != nil {
			logger.Error(err)
			continue
		}
		s.masterDo(election, nodeName)
	}
}

func (s *Scheduler) masterDo(election *concurrency.Election, nodeName string) {
	stopMasterCh := make(chan int)
	go s.watchApiKeyTask(stopMasterCh)
	go s.nodeCacheTask(stopMasterCh)
	go s.watchStrategyRegisterTask(stopMasterCh)

	observeCh := election.Observe(context.Background())
	for {
		select {
		case <-s.closed:
			logger.Info("masterDo exit")
			return

		case resp, ok := <-observeCh:
			if !ok {
				logger.Error("election observe closed")
				close(stopMasterCh)
				return
			}

			logger.Warn("electron:", resp)
			if string(resp.Kvs[0].Value) != nodeName {
				logger.Error("current node is not master")
				close(stopMasterCh)
				return
			}
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

// 寻找运行策略最少的节点
func (s *Scheduler) getAppropriateNode(strategy string) string {
	var addr string
	prefix := EtcdStrategyNodePrefix + strategy + "/"

	minCount := math.MaxInt64
	s.nodes.Range(func(key, value interface{}) bool {
		path := key.(string)
		if strings.HasPrefix(path, prefix) {
			ni, ok := value.(*NodeInfo)
			if !ok {
				logger.Error("node cache parse value fail")
				return false
			}

			if len(ni.ApiKeys) < minCount {
				minCount = len(ni.ApiKeys)
				addr = path
			}
		}
		return true
	})
	if addr == "" {
		return ""
	}
	return strings.TrimPrefix(addr, prefix)
}

func (s *Scheduler) nodeCacheTask(stopMasterCh chan int) {
	for {
		select {
		case <-s.closed:
			logger.Info("nodeCacheTask exit")
			return
		case <-stopMasterCh:
			logger.Info("nodeCacheTask exit")
			return
		default:
		}

		s.nodeCache(stopMasterCh)
	}
}

func (s *Scheduler) nodeCache(stopMasterCh chan int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.nodes.Range(func(key, value interface{}) bool {
		s.nodes.Delete(key)
		return true
	})

	// /scheduler/strategy/strategyName/10.22.33.55:32154 --> {"apiKeys":["apikey1","apikey2"]}
	getResp, err := s.client.Get(ctx, EtcdStrategyNodePrefix, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return
	}
	for _, kv := range getResp.Kvs {
		ni := &NodeInfo{}
		if err := json.Unmarshal(kv.Value, ni); err != nil {
			logger.Error(err)
			continue
		}
		s.nodes.Store(string(kv.Key), ni)
	}
	rch := s.client.Watch(ctx, EtcdStrategyNodePrefix, clientv3.WithPrefix())

	for {
		select {
		case wresp, ok := <-rch:
			if !ok {
				logger.Errorf("watch %s exit", EtcdStrategyNodePrefix)
				return
			}

			for _, ev := range wresp.Events {
				logger.Infof("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)

				switch ev.Type {
				case mvccpb.PUT:
					ni := &NodeInfo{}
					if err := json.Unmarshal(ev.Kv.Value, ni); err != nil {
						logger.Error(err)
						break
					}
					s.nodes.Store(string(ev.Kv.Key), ni)
				case mvccpb.DELETE:
					s.nodes.Delete(string(ev.Kv.Key))
				}
			}

		case <-s.closed:
			return
		case <-stopMasterCh:
			return
		}
	}
}

func (s *Scheduler) Close() {
	s.once.Do(func() {
		close(s.closed)
		s.client.Close()
	})
}
