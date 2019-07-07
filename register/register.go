package register

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/harveywangdao/ants/logger"
	"github.com/harveywangdao/ants/util"
	"net"
	"strings"
	"time"
)

const (
	dialTimeout          = 5 * time.Second
	requestTimeout       = 10 * time.Second
	leaseTimeout         = 4 //秒
	serverRegisterPrefix = "/ants/services/"
)

type Register struct {
	serviceInfo *ServiceInfo
	cli         *clientv3.Client
	stop        chan struct{}
	leaseid     clientv3.LeaseID
	endpoints   []string
}

type ServiceInfo struct {
	Ip           string    `json:"ip"`
	Port         string    `json:"port"`
	Name         string    `json:"name"`
	RegisterTime time.Time `json:"registerTime"`
}

func GetLocalIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Error(err)
		return ""
	}

	ip := ""
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				logger.Info("ip:", ip)
			}
		}
	}

	return ip
}

func GetLocalIp2() string {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		logger.Error(err)
		return ""
	}

	ip := ""

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						ip = ipnet.IP.String()
						logger.Info("ip:", ip)
					}
				}
			}
		}
	}

	return ip
}

func GetLocalIp3() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logger.Error(err)
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().String()
	logger.Debug("localAddr", localAddr)
	idx := strings.LastIndex(localAddr, ":")

	localIp := localAddr[0:idx]
	logger.Info("local ip:", localIp)

	return localIp
}

func NewRegister(endpoints []string, port string, name string) (*Register, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	r := new(Register)
	r.cli = cli

	s := &ServiceInfo{
		Ip:           GetLocalIp3(),
		Port:         port,
		Name:         name,
		RegisterTime: time.Now(),
	}

	r.serviceInfo = s
	r.stop = make(chan struct{})
	r.endpoints = endpoints

	return r, nil
}

func (r *Register) Start() {
	go func() {
		defer r.cli.Close()

		defer func() {
			if r := recover(); r != nil {
				logger.Error(r)
			}
		}()

		if err := r.leaseKey(); err != nil {
			return
		}

		for {
			select {
			case <-r.stop:
				_, err := r.cli.Revoke(context.TODO(), r.leaseid)
				if err != nil {
					logger.Error(err)
					return
				}

				logger.Info("etcd register stop")
				return
			case <-time.After((leaseTimeout - 1) * time.Second):
				ch, err := r.cli.KeepAlive(context.TODO(), r.leaseid)
				if err != nil {
					logger.Error(err)
					return
				}

				ka, ok := <-ch
				if !ok {
					logger.Error("KeepAlive fail")
					//1.连接断开
					//2.租期已过
					if !r.leaseExist() {
						if err := r.leaseKey(); err != nil {
							return
						}
					}

					continue
				}

				logger.Debug("ttl:", ka.TTL)
			}
		}
	}()
}

func (r *Register) leaseExist() bool {
	timeToLiveResponse, err := r.cli.TimeToLive(context.TODO(), r.leaseid)
	if err != nil {
		logger.Error(err)
		return false
	}

	data, _ := json.Marshal(timeToLiveResponse)
	logger.Info(string(data))

	if timeToLiveResponse.TTL == -1 {
		return false
	}

	return true
}

func (r *Register) leaseKey() error {
	resp, err := r.cli.Grant(context.TODO(), leaseTimeout)
	if err != nil {
		logger.Error(err)
		return err
	}

	r.leaseid = resp.ID

	data, _ := json.Marshal(r.serviceInfo)
	key := serverRegisterPrefix + r.serviceInfo.Name + "/" + util.GetUUID()
	_, err = r.cli.Put(context.TODO(), key, string(data), clientv3.WithLease(r.leaseid))
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (r *Register) Stop() {
	r.stop <- struct{}{}
}
