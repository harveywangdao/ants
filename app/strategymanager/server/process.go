package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/harveywangdao/ants/app/scheduler/util"
	"github.com/harveywangdao/ants/app/scheduler/util/logger"
	spb "github.com/harveywangdao/ants/app/strategymanager/protos/strategy"
	mgrpb "github.com/harveywangdao/ants/app/strategymanager/protos/strategymanager"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	mvccpb "go.etcd.io/etcd/mvcc/mvccpb"
	"google.golang.org/grpc"
)

var (
	// 1:进程未启动;2:进程已启动还未start;3.start成功
	StateProcessStop   = int32(1)
	StateProcessStart  = int32(2)
	StateStrategyStart = int32(3)
)

/*
资源:
1.process
2.monitorConn
3.grpcConn
*/
type StrategyProcess struct {
	mgr   *StrategyManager
	state int32

	uniqueId         string
	process          *os.Process
	monitorConn      net.Conn
	restartProcessCh chan bool

	unixFile string
	grpcConn *grpc.ClientConn
	Client   spb.StrategyClient

	startStrategyReq *spb.StartStrategyRequest
	startCh          chan bool

	once    sync.Once
	closeCh chan bool

	leaseid clientv3.LeaseID
}

func NewStrategyProcess(mgr *StrategyManager, startReq *mgrpb.StartTaskRequest) (*StrategyProcess, error) {
	unixFile := filepath.Join(os.TempDir(), util.GetUUID())
	sp := &StrategyProcess{
		mgr:              mgr,
		uniqueId:         fmt.Sprintf("%s-%s-%s", startReq.ApiKey, startReq.StrategyName, startReq.InstrumentId),
		unixFile:         unixFile,
		restartProcessCh: make(chan bool, 1),
		startCh:          make(chan bool, 1),
		closeCh:          make(chan bool),
		state:            StateProcessStop,
	}

	// 创建无名unix domain socket,用来监控子进程是否挂掉
	fds, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM, 0)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	f1 := os.NewFile(uintptr(fds[0]), "f1")
	f2 := os.NewFile(uintptr(fds[1]), "f2")
	defer f1.Close()
	defer f2.Close()

	c1, err := net.FileConn(f1)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer c1.Close()

	c2, err := net.FileConn(f2)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// 创建进程
	path := filepath.Join(mgr.getProcDir(), startReq.StrategyName)
	argv := []string{path, unixFile}
	attr := &os.ProcAttr{
		//Env: os.Environ(),
		Files: []*os.File{nil, nil, nil, f1},
	}
	process, err := os.StartProcess(path, argv, attr)
	if err != nil {
		logger.Error(err)
		c2.Close()
		return nil, err
	}

	// 连接grpc
	conn, err := grpc.Dial(unixFile, grpc.WithInsecure(), grpc.WithDialer(sp.unixConnect))
	if err != nil {
		logger.Error(err)
		c2.Close()
		process.Kill()
		process.Wait()
		return nil, err
	}

	sp.monitorConn = c2
	sp.process = process
	sp.grpcConn = conn
	sp.Client = spb.NewStrategyClient(sp.grpcConn)

	sp.startStrategyReq = &spb.StartStrategyRequest{
		UserId:       startReq.UserId,
		Exchange:     startReq.Exchange,
		ApiKey:       startReq.ApiKey,
		SecretKey:    startReq.SecretKey,
		Passphrase:   startReq.Passphrase,
		StrategyName: startReq.StrategyName,
		InstrumentId: startReq.InstrumentId,
		Endpoint:     startReq.Endpoint,
		WsEndpoint:   startReq.WsEndpoint,
		Params:       startReq.Params,
	}

	sp.state = StateProcessStart
	sp.startCh <- true

	go sp.task()
	go sp.watchTask()

	return sp, nil
}

func (sp *StrategyProcess) task() {
	tk := time.NewTicker(time.Second * 3)

	defer func() {
		sp.destoryProcess()

		if err := sp.grpcConn.Close(); err != nil {
			logger.Error(err)
		}

		tk.Stop()
		sp.mgr.delProcess(sp.uniqueId)
		// 删除etcd注册信息
	}()

	for {
		select {
		case <-sp.closeCh:
			return

		case <-tk.C:
			state := atomic.LoadInt32(&sp.state)
			if state == StateStrategyStart {
				// 向etcd注册
				// 刷新etcd
			}

		case <-sp.startCh:
			// 开始策略,由于策略进程可能还没正真启动,第一次可能会失败,需要重试
			if err := sp.startStrategy(); err != nil {
				logger.Error(err)

				state := atomic.LoadInt32(&sp.state)
				if state == StateProcessStart {
					sp.startCh <- true
				}

				time.Sleep(time.Second)
			} else {
				atomic.StoreInt32(&sp.state, StateStrategyStart)
			}
		case <-sp.restartProcessCh:
			atomic.StoreInt32(&sp.state, StateProcessStop)
			sp.destoryProcess()
			time.Sleep(time.Second)
			if err := sp.restartProcess(); err != nil {
				logger.Error("restart process fail, err:", err)
				sp.restartProcessCh <- true
			} else {
				go sp.watchTask()
				atomic.StoreInt32(&sp.state, StateProcessStart)
				sp.startCh <- true
			}
		}
	}
}

func (sp *StrategyProcess) watchTask() {
	buf := make([]byte, 32)
	// 一直阻塞直到策略进程退出
	_, err := sp.monitorConn.Read(buf)
	select {
	case <-sp.closeCh:
		return
	default:
	}

	logger.Error(err)
	sp.restartProcessCh <- true
}

type StrategyNode struct {
	Addr string `json:"addr"`
}

func (s *StrategyProcess) register() error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: s.mgr.Config.Etcd.Endpoints,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	defer cli.Close()

	if s.leaseid != 0 {
		timeToLiveResp, err := cli.TimeToLive(context.TODO(), s.leaseid)
		if err != nil {
			logger.Error(err)
			return err
		}
		if timeToLiveResp.TTL == -1 {
			s.leaseid = 0
		}
	}

	if s.leaseid == 0 {
		resp, err := cli.Grant(context.TODO(), 5*time.Second)
		if err != nil {
			logger.Error(err)
			return err
		}
		s.leaseid = resp.ID

		node := &StrategyNode{
			Addr: "",
		}
		data, err := json.Marshal(node)
		if err != nil {
			logger.Error(err)
			return err
		}

		key := "/strategy/"
		_, err = cli.Put(context.TODO(), key, string(data), clientv3.WithLease(s.leaseid))
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	resp, err := cli.KeepAliveOnce(context.TODO(), s.leaseid)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info(resp)

	return nil
}

func (s *StrategyProcess) unregister() error {
	if s.leaseid == 0 {
		return nil
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints: s.mgr.Config.Etcd.Endpoints,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	defer cli.Close()

	_, err = cli.Revoke(context.TODO(), s.leaseid)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (sp *StrategyProcess) restartProcess() error {
	// 创建无名unix domain socket,用来监控子进程是否挂掉
	fds, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM, 0)
	if err != nil {
		logger.Error(err)
		return err
	}
	f1 := os.NewFile(uintptr(fds[0]), "f1")
	f2 := os.NewFile(uintptr(fds[1]), "f2")
	defer f1.Close()
	defer f2.Close()

	c1, err := net.FileConn(f1)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer c1.Close()

	c2, err := net.FileConn(f2)
	if err != nil {
		logger.Error(err)
		return err
	}

	// 创建进程
	path := filepath.Join(sp.mgr.getProcDir(), sp.startStrategyReq.StrategyName)
	argv := []string{path, sp.unixFile}
	attr := &os.ProcAttr{
		Files: []*os.File{nil, nil, nil, f1},
	}
	process, err := os.StartProcess(path, argv, attr)
	if err != nil {
		logger.Error(err)
		c2.Close()
		return err
	}

	sp.monitorConn = c2
	sp.process = process

	return nil
}

func (sp *StrategyProcess) destoryProcess() {
	if err := sp.monitorConn.Close(); err != nil {
		logger.Error(err)
	}

	if err := os.RemoveAll(sp.unixFile); err != nil {
		logger.Error(err)
	}

	if err := sp.process.Kill(); err != nil {
		logger.Error(err)
	}
	procState, err := sp.process.Wait()
	if err != nil {
		logger.Error(err)
	}
	logger.Infof("proc id: %s exit state: %s", procState.Pid(), procState.String())
}

func (sp *StrategyProcess) startStrategy() error {
	_, err = sp.Client.StartStrategy(ctx, sp.startStrategyReq)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (sp *StrategyProcess) unixConnect(addr string, t time.Duration) (net.Conn, error) {
	raddr, err := net.ResolveUnixAddr("unix", sp.unixFile)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return net.DialUnix("unix", nil, raddr)
}

func (sp *StrategyProcess) Close() {
	sp.once.Do(func() {
		close(sp.closeCh)
	})
}

/*
1.更新参数
2.etcd注册，刷新
3.如何获取本地地址
*/
