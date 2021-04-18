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

	"github.com/gin-gonic/gin"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	mvccpb "go.etcd.io/etcd/mvcc/mvccpb"
)

const (
	EtcdStrategyElectionKey  = "/scheduler/master"
	EtcdSchedulerNodePrefix  = "/scheduler/node/"
	EtcdServiceManagerPrefix = "/service/strategymanager/"
)

type Scheduler struct {
	client *clientv3.Client

	startCh chan *StrategyData
	stopCh  chan *StrategyData

	closed chan bool
	once   sync.Once

	nodeTaskCount atomic.Value
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
		startCh: make(chan *StrategyData),
		stopCh:  make(chan *StrategyData),
		closed:  make(chan bool),
	}
	s.nodeTaskCount.Store(make(map[string]int))

	go s.masterRun()
	go s.nodeCacheTask()
	return s, nil
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
	getResp, err := s.client.Get(ctx, EtcdServiceManagerPrefix, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return
	}

	minCount := math.MaxInt64
	var addr string
	for _, kv := range getResp.Kvs {
		logger.Info("node key:", string(kv.Key))
		nodeAddr := strings.TrimPrefix(string(kv.Key), EtcdServiceManagerPrefix)
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
	getResp, err := s.client.Get(ctx, EtcdSchedulerNodePrefix, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return
	}
	for _, kv := range getResp.Kvs {
		logger.Info("node key:", string(kv.Key))
		parts := strings.Split(strings.TrimPrefix(string(kv.Key), EtcdSchedulerNodePrefix), "/")
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
	rch := s.client.Watch(ctx, EtcdSchedulerNodePrefix, clientv3.WithPrefix())

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

				parts := strings.Split(strings.TrimPrefix(string(ev.Kv.Key), EtcdSchedulerNodePrefix), "/")
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
	session, err := concurrency.NewSession(s.client)
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
		s.client.Close()
	})
}
