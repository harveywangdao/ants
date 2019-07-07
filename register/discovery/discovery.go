package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/coreos/etcd/clientv3"
	"github.com/harveywangdao/ants/logger"
	"github.com/harveywangdao/ants/register"
	"time"
)

const (
	serverRegisterPrefix = "/ants/services/"
	dialTimeout          = 5 * time.Second
)

type Discovery struct {
	cli       *clientv3.Client
	endpoints []string
}

func NewDiscovery(endpoints []string) (*Discovery, error) {
	d := new(Discovery)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	d.cli = cli
	d.endpoints = endpoints

	return d, nil
}

func (d *Discovery) QueryServiceIpPort(name string) ([]string, error) {
	getResp, err := d.cli.Get(context.TODO(), serverRegisterPrefix+name, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(getResp.Kvs) == 0 {
		logger.Error(name, "not existed")
		return nil, errors.New(name + " not existed")
	}

	var addrs []string
	for _, ev := range getResp.Kvs {
		logger.Infof("%s : %s\n", ev.Key, ev.Value)

		var serviceInfo register.ServiceInfo
		err = json.Unmarshal(ev.Value, &serviceInfo)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		addrs = append(addrs, serviceInfo.Ip+":"+serviceInfo.Port)
	}

	return addrs, nil
}

func (d *Discovery) get(key string) {
	getResp, err := d.cli.Get(context.TODO(), key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		logger.Error(err)
		return
	}

	if len(getResp.Kvs) == 0 {
		logger.Error(key, "not existed")
		return
	}

	for _, ev := range getResp.Kvs {
		logger.Infof("%s : %s\n", ev.Key, ev.Value)
	}
}

func (d *Discovery) GetAllService() {
	d.get(serverRegisterPrefix)
}
