package server

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/harveywangdao/ants/app/scheduler/model"
	"github.com/harveywangdao/ants/app/scheduler/util"
	"github.com/harveywangdao/ants/app/scheduler/util/logger"
	mgrpb "github.com/harveywangdao/ants/app/strategymanager/protos/strategymanager"

	"github.com/gin-gonic/gin"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/clientv3util"
	"go.etcd.io/etcd/clientv3/concurrency"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"google.golang.org/grpc"
)

const (
	EtcdStrategyElectionKey     = "/scheduler/master"
	EtcdSchedulerNodePrefix     = "/scheduler/node"
	EtcdServiceManagerPrefix    = "/service/strategymanager"
	EtcdSchedulerStrategyPrefix = "/scheduler/strategy"
)

type Scheduler struct {
	etcdClient *clientv3.Client

	startCh chan *StrategyData
	stopCh  chan *StrategyData

	closed chan bool
	once   sync.Once

	nodeTaskCount atomic.Value
}

type SchedulerNodeInfo struct {
	Uptime int64 `json:"uptime"`
}

type StrategyInfo struct {
	Uptime    int64  `json:"uptime"`
	Available bool   `json:"available"`
	Addr      string `json:"addr"`
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
		etcdClient: cli,
		startCh:    make(chan *StrategyData),
		stopCh:     make(chan *StrategyData),
		closed:     make(chan bool),
	}
	s.nodeTaskCount.Store(make(map[string]int))

	go s.masterRun()
	go s.nodeCacheTask()
	return s, nil
}

func (s *Scheduler) StartOneStrategyTask(req *StrategyData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodeAddr := s.getAppropriateNode()
	if nodeAddr == "" {
		return fmt.Errorf("strategy node unavailable")
	}

	// /strategy/task/$apikey/$strategy/$instrumentid
	strategyTaskPath := fmt.Sprintf("%s/%s/%s/%s", StrategyTaskPrefix, req.ApiKey, req.StrategyName, req.Symbol)

	// /scheduler/node/10.22.33.55:32154/$apikey/$strategy/$instrumentid --> {"uptime":111111111}
	schedulerNodePath := fmt.Sprintf("%s/%s/%s/%s/%s", EtcdSchedulerNodePrefix, nodeAddr, req.ApiKey, req.StrategyName, req.Symbol)
	schedulerNodeInfoData, err := json.Marshal(&SchedulerNodeInfo{
		Uptime: time.Now().Unix(),
	})
	if err != nil {
		logger.Error(err)
		return err
	}

	// /scheduler/strategy/$apikey/$strategy/$instrumentid --> {"addr":"10.22.33.55:32154", "uptime":111111111, "available":true}
	schedulerStrategyPath := fmt.Sprintf("%s/%s/%s/%s", EtcdSchedulerStrategyPrefix, req.ApiKey, req.StrategyName, req.Symbol)
	strategyInfoData, err := json.Marshal(&StrategyInfo{
		Uptime:    time.Now().Unix(),
		Available: true,
		Addr:      nodeAddr,
	})
	if err != nil {
		logger.Error(err)
		return err
	}

	// 获取参数
	getResp, err := s.etcdClient.Get(ctx, strategyTaskPath)
	if err != nil {
		logger.Error(err)
		return err
	}
	if len(getResp.Kvs) != 1 {
		return fmt.Errorf("%s error: %v", strategyTaskPath, getResp)
	}
	strategyTaskInfo := &StrategyTaskInfo{}
	if err := json.Unmarshal(getResp.Kvs[0].Value, strategyTaskInfo); err != nil {
		logger.Error(err)
		return err
	}

	// 检查策略是否存在
	// 检查策略对应关系是否存在
	// etcd增加映射信息
	kvc := clientv3.NewKV(s.etcdClient)
	txnResp, err := kvc.Txn(ctx).
		If(clientv3util.KeyExists(strategyTaskPath), clientv3util.KeyMissing(schedulerNodePath), clientv3util.KeyMissing(schedulerStrategyPath)).
		Then(clientv3.OpPut(schedulerNodePath, string(schedulerNodeInfoData)), clientv3.OpPut(schedulerStrategyPath, string(strategyInfoData))).
		Commit()
	if err != nil {
		logger.Error(err)
		return err
	}
	if !txnResp.Succeeded {
		return fmt.Errorf("%s not existed, %s or %s already existed", strategyTaskPath, schedulerNodePath, schedulerStrategyPath)
	}

	revoke := func() error {
		txnResp2, err := kvc.Txn(ctx).If().
			Then(clientv3.OpDelete(schedulerNodePath), clientv3.OpDelete(schedulerStrategyPath)).
			Commit()
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	// 连接策略服务
	conn, err := grpc.DialContext(ctx, nodeAddr, grpc.WithInsecure())
	if err != nil {
		logger.Error(err)
		revoke()
		return err
	}
	mgrCli := mgrpb.NewStrategyManagerClient(conn)

	// 开始策略
	startTaskReq := &mgrpb.StartTaskRequest{
		UserId:       strategyTaskInfo.MetaData.UserId,
		Exchange:     strategyTaskInfo.MetaData.Exchange,
		ApiKey:       strategyTaskInfo.MetaData.ApiKey,
		SecretKey:    strategyTaskInfo.MetaData.SecretKey,
		Passphrase:   strategyTaskInfo.MetaData.Passphrase,
		StrategyName: strategyTaskInfo.MetaData.StrategyName,
		InstrumentId: strategyTaskInfo.MetaData.Symbol,
		Params:       strategyTaskInfo.MetaData.Params,
	}
	if _, err := mgrCli.StartTask(ctx, startTaskReq); err != nil {
		logger.Error(err)
		revoke()
		return err
	}

	return nil
}

func (s *Scheduler) IsTaskRunning(req *StrategyData) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// /scheduler/strategy/$apikey/$strategy/$instrumentid --> {"addr":"10.22.33.55:32154", "uptime":111111111, "available":true}
	schedulerStrategyPath := fmt.Sprintf("%s/%s/%s/%s", EtcdSchedulerStrategyPrefix, req.ApiKey, req.StrategyName, req.Symbol)

	// 查询策略节点地址
	getResp, err := s.etcdClient.Get(ctx, schedulerStrategyPath)
	if err != nil {
		logger.Error(err)
		return false, err
	}
	if len(getResp.Kvs) == 0 {
		return false, nil
	}
	return true, nil
}

func (s *Scheduler) StopOneStrategyTask(req *StrategyData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// /scheduler/strategy/$apikey/$strategy/$instrumentid --> {"addr":"10.22.33.55:32154", "uptime":111111111, "available":true}
	schedulerStrategyPath := fmt.Sprintf("%s/%s/%s/%s", EtcdSchedulerStrategyPrefix, req.ApiKey, req.StrategyName, req.Symbol)

	// 查询策略节点地址
	getResp, err := s.etcdClient.Get(ctx, schedulerStrategyPath)
	if err != nil {
		logger.Error(err)
		return err
	}
	if len(getResp.Kvs) == 0 {
		logger.Errorf("%s not existed", schedulerStrategyPath)
		return nil
	}
	strategyInfo := &StrategyInfo{}
	if err := json.Unmarshal(getResp.Kvs[0].Value, strategyInfo); err != nil {
		logger.Error(err)
		return err
	}

	// 连接策略服务
	conn, err := grpc.DialContext(ctx, strategyInfo.Addr, grpc.WithInsecure())
	if err != nil {
		logger.Error(err)
		return err
	}
	mgrCli := mgrpb.NewStrategyManagerClient(conn)

	// 停止策略
	stopTaskReq := &mgrpb.StopTaskRequest{
		ApiKey:       req.ApiKey,
		StrategyName: req.StrategyName,
		InstrumentId: req.Symbol,
	}
	if _, err := mgrCli.StopTask(ctx, stopTaskReq); err != nil {
		logger.Error(err)
		return err
	}

	// etcd删除映射信息
	// /scheduler/node/10.22.33.55:32154/$apikey/$strategy/$instrumentid --> {"uptime":111111111}
	schedulerNodePath := fmt.Sprintf("%s/%s/%s/%s/%s", EtcdSchedulerNodePrefix, strategyInfo.Addr, req.ApiKey, req.StrategyName, req.Symbol)

	kvc := clientv3.NewKV(s.etcdClient)
	txnResp, err := kvc.Txn(ctx).If().
		Then(clientv3.OpDelete(schedulerNodePath), clientv3.OpDelete(schedulerStrategyPath)).
		Commit()
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (s *Scheduler) UpdateOneStrategyTask(req *StrategyData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// /scheduler/strategy/$apikey/$strategy/$instrumentid --> {"addr":"10.22.33.55:32154", "uptime":111111111, "available":true}
	schedulerStrategyPath := fmt.Sprintf("%s/%s/%s/%s", EtcdSchedulerStrategyPrefix, req.ApiKey, req.StrategyName, req.Symbol)

	// 查询策略节点地址
	getResp, err := s.etcdClient.Get(ctx, schedulerStrategyPath)
	if err != nil {
		logger.Error(err)
		return err
	}
	if len(getResp.Kvs) == 0 {
		err = fmt.Errorf("%s not existed", schedulerStrategyPath)
		logger.Errorf(err)
		return err
	}
	strategyInfo := &StrategyInfo{}
	if err := json.Unmarshal(getResp.Kvs[0].Value, strategyInfo); err != nil {
		logger.Error(err)
		return err
	}

	// 连接策略服务
	conn, err := grpc.DialContext(ctx, strategyInfo.Addr, grpc.WithInsecure())
	if err != nil {
		logger.Error(err)
		return err
	}
	mgrCli := mgrpb.NewStrategyManagerClient(conn)

	// 更新策略参数
	execReq := &mgrpb.TaskCommandExecRequest{
		ApiKey:       req.ApiKey,
		StrategyName: req.StrategyName,
		InstrumentId: req.Symbol,
		Params:       req.Params,
	}
	if _, err := mgrCli.TaskCommandExec(ctx, execReq); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

// 寻找运行策略最少的节点
func (s *Scheduler) getAppropriateNode() string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodeTaskCountMap := s.nodeTaskCount.Load().(map[string]int)
	if nodeTaskCountMap == nil {
		logger.Error("nodeTaskCount is nil")
		return ""
	}

	// 获取所有当前在线的节点
	// 遍历哪个节点运行的任务最少
	// /service/strategymanager/10.22.33.55:32154 --> {"uptime":111111111, "available":true}
	getResp, err := s.etcdClient.Get(ctx, EtcdServiceManagerPrefix, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return
	}

	minCount := math.MaxInt64
	var addr string
	for _, kv := range getResp.Kvs {
		logger.Info("node key:", string(kv.Key))
		nodeAddr := strings.TrimPrefix(string(kv.Key), EtcdServiceManagerPrefix+"/")
		count := nodeTaskCountMap[nodeAddr]
		if count < minCount {
			addr = nodeAddr
			minCount = count
			if minCount == 0 {
				break
			}
		}
	}
	return addr
}

func (s *Scheduler) nodeCacheTask() {
	for {
		select {
		case <-s.closed:
			return
		default:
		}
		s.nodeCache()
	}
}

func (s *Scheduler) nodeCache() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodesTable := make(map[string]map[string]int)

	// /scheduler/node/10.22.33.55:32154/$apikey/$strategy/$instrumentid
	getResp, err := s.etcdClient.Get(ctx, EtcdSchedulerNodePrefix, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return
	}
	for _, kv := range getResp.Kvs {
		logger.Info("node key:", string(kv.Key))
		parts := strings.Split(strings.TrimPrefix(string(kv.Key), EtcdSchedulerNodePrefix+"/"), "/")
		if len(parts) != 4 {
			logger.Error("node key error, key:", string(kv.Key))
			continue
		}
		nodeKey := parts[0]
		taskKey := fmt.Sprintf("%s/%s/%s", parts[1], parts[2], parts[3])

		tasks, ok := nodesTable[nodeKey]
		if !ok {
			tasks = make(map[string]int)
		}
		tasks[taskKey] = 1
		nodesTable[nodeKey] = tasks
	}
	rch := s.etcdClient.Watch(ctx, EtcdSchedulerNodePrefix, clientv3.WithPrefix())

	tk := time.NewTicker(time.Second * 3)
	defer tk.Stop()

	for {
		select {
		case wresp, ok := <-rch:
			if !ok {
				logger.Errorf("watch %s exit", EtcdSchedulerNodePrefix)
				return
			}

			for _, ev := range wresp.Events {
				logger.Infof("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)

				parts := strings.Split(strings.TrimPrefix(string(ev.Kv.Key), EtcdSchedulerNodePrefix+"/"), "/")
				if len(parts) != 4 {
					logger.Error("node key error, key:", string(ev.Kv.Key))
					continue
				}
				nodeKey := parts[0]
				taskKey := fmt.Sprintf("%s/%s/%s", parts[1], parts[2], parts[3])

				switch ev.Type {
				case mvccpb.PUT:
					tasks, ok := nodesTable[nodeKey]
					if !ok {
						tasks = make(map[string]int)
					}
					tasks[taskKey] = 1
					nodesTable[nodeKey] = tasks
				case mvccpb.DELETE:
					if tasks, ok := nodesTable[nodeKey]; ok {
						delete(tasks, taskKey)
						if len(tasks) == 0 {
							delete(nodesTable, nodeKey)
						}
					}
				}
			}

		case <-tk.C:
			nodes := make(map[string]int)
			for nodeKey, tasks := range nodesTable {
				nodes[nodeKey] = len(tasks)
			}
			s.nodeTaskCount.Store(nodes)
		case <-s.closed:
			return
		}
	}
}

// 一主多从结构,使用etcd的选举
func (s *Scheduler) masterRun() {
	session, err := concurrency.NewSession(s.etcdClient)
	if err != nil {
		logger.Fatal(err)
		return
	}
	election := concurrency.NewElection(session, EtcdStrategyElectionKey)
	nodeName := util.GetUUID()
	logger.Info("current node name:", nodeName)

	for {
		select {
		case <-s.closed:
			return
		default:
		}

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

func (s *Scheduler) Close() {
	s.once.Do(func() {
		close(s.closed)
		s.etcdClient.Close()
	})
}
