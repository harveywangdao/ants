package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"ants/logger"
	"github.com/coreos/etcd/clientv3"
)

var (
	endpoints      = []string{"192.168.1.7:2379", "192.168.1.10:2379", "192.168.1.11:2379"}
	dialTimeout    = 5 * time.Second
	requestTimeout = 10 * time.Second
)

func init() {
	//fileHandler := logger.NewFileHandler("test.log")
	//logger.SetHandlers(logger.Console, fileHandler)
	logger.SetHandlers(logger.Console)
	//defer logger.Close()
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	logger.SetLevel(logger.INFO)
}

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		logger.Error(err)
		return
	}
	defer cli.Close()

	//put
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	putResp, err := cli.Put(ctx, "sample_key", "sample_value")
	cancel()
	if err != nil {
		logger.Error(err)
		return
	}
	data, _ := json.Marshal(putResp)
	logger.Info(string(data))

	//get
	ctx, cancel = context.WithTimeout(context.Background(), requestTimeout)
	getResp, err := cli.Get(ctx, "sample_key")
	cancel()
	if err != nil {
		logger.Error(err)
		return
	}
	data, _ = json.Marshal(getResp)
	logger.Info(string(data))

	for _, ev := range getResp.Kvs {
		logger.Infof("%s : %s\n", ev.Key, ev.Value)
	}

	//get prefix
	cli.Put(context.TODO(), "sample_key1", "sample_value")
	cli.Put(context.TODO(), "sample_key2", "sample_value")
	cli.Put(context.TODO(), "sample_key3", "sample_value")

	ctx, cancel = context.WithTimeout(context.Background(), requestTimeout)
	getResp, err = cli.Get(ctx, "sample_key", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	cancel()
	if err != nil {
		logger.Error(err)
		return
	}
	data, _ = json.Marshal(getResp)
	logger.Info(string(data))

	for _, ev := range getResp.Kvs {
		logger.Infof("%s : %s\n", ev.Key, ev.Value)
	}

	//delete
	ctx, cancel = context.WithTimeout(context.Background(), requestTimeout)
	delResp, err := cli.Delete(ctx, "sample_key", clientv3.WithPrefix())
	cancel()
	if err != nil {
		logger.Error(err)
		return
	}
	data, _ = json.Marshal(delResp)
	logger.Info(string(data))

	getResp, err = cli.Get(context.TODO(), "sample_key", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		logger.Error(err)
		return
	}
	data, _ = json.Marshal(getResp)
	logger.Info(string(data))

	//txn
	txn_test()

	//lease
	lease_test()
}

func txn_test() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		logger.Error(err)
		return
	}
	defer cli.Close()

	kvc := clientv3.NewKV(cli)

	_, err = kvc.Put(context.TODO(), "key", "xyz")
	if err != nil {
		logger.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err = kvc.Txn(ctx).
		// txn value comparisons are lexical
		If(clientv3.Compare(clientv3.Value("key"), ">", "abc")).
		// the "Then" runs, since "xyz" > "abc"
		Then(clientv3.OpPut("key", "XYZ")).
		// the "Else" does not run
		Else(clientv3.OpPut("key", "ABC")).
		Commit()
	cancel()
	if err != nil {
		logger.Error(err)
		return
	}

	getResp, err := kvc.Get(context.TODO(), "key")
	cancel()
	if err != nil {
		logger.Error(err)
		return
	}
	for _, ev := range getResp.Kvs {
		logger.Infof("%s : %s\n", ev.Key, ev.Value)
	}
	// Output: key : XYZ
}

func getValue(cli *clientv3.Client, key string) {
	getResp, err := cli.Get(context.TODO(), key)
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

func lease_test() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		logger.Error(err)
		return
	}
	defer cli.Close()

	resp, err := cli.Grant(context.TODO(), 4)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("resp.ID:", resp.ID)

	resp2, err := cli.Grant(context.TODO(), 40)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("resp2.ID:", resp2.ID)

	/*_, err = cli.Revoke(context.TODO(), resp.ID)
	if err != nil {
		logger.Error(err)
		return
	}*/

	timeToLiveResponse, err := cli.TimeToLive(context.TODO(), resp.ID)
	if err != nil {
		logger.Error(err)
		return
	}
	data, _ := json.Marshal(timeToLiveResponse)
	logger.Info(string(data))

	leasesResponse, err := cli.Leases(context.TODO())
	if err != nil {
		logger.Error(err)
		return
	}
	data, _ = json.Marshal(leasesResponse)
	logger.Info(string(data))

	_, err = cli.Put(context.TODO(), "lease_key", "lease_value", clientv3.WithLease(resp.ID))
	if err != nil {
		logger.Error(err)
		return
	}
	getValue(cli, "lease_key")

	time.Sleep(3 * time.Second)

	timeToLiveResponse, err = cli.TimeToLive(context.TODO(), resp.ID)
	if err != nil {
		logger.Error(err)
		return
	}
	data, _ = json.Marshal(timeToLiveResponse)
	logger.Info(string(data))

	ch, err := cli.KeepAlive(context.TODO(), resp.ID)
	if err != nil {
		logger.Error(err)
		return
	}
	ka := <-ch
	logger.Info("ttl:", ka.TTL)

	time.Sleep(3 * time.Second)

	getValue(cli, "lease_key")

	_, err = cli.Put(context.TODO(), "lease_key2", "lease_value2", clientv3.WithLease(resp.ID))
	if err != nil {
		logger.Error(err)
	}
}
