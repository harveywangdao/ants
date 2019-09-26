package main

import (
	"fmt"
	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
	"reflect"
	"time"
)

func do1() {
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", "192.168.1.7:7001") },
	}

	conn := pool.Get()
	defer conn.Close()

	reply, err := conn.Do("SET", "DeductStock", "46554621635", "NX", "EX", 10)
	if err != nil {
		fmt.Println(err)
		return
	}

	setOk, ok := reply.(string)
	if !ok {
		fmt.Println("reply:", reply)
		return
	}

	fmt.Println("setOk:", setOk)
	fmt.Println(setOk == "OK")

	// OK <nil>
	// <nil> <nil>
	time.Sleep(1 * time.Second)

	sc := redis.NewScript(1, `
	if redis.call("get",KEYS[1]) == ARGV[1]
	then
	    return redis.call("del",KEYS[1])
	else
	    return 0
	end`)

	reply, err = sc.Do(conn, "DeductStock", "46554621635")
	if err != nil {
		fmt.Println(err)
		return
	}

	// reply: 1
	// reply: 0
	fmt.Println("reply:", reply)
	fmt.Println("reply:", reflect.TypeOf(reply))

	replyStr := fmt.Sprintf("%v", reply)
	fmt.Println("replyStr:", replyStr)
	fmt.Println("replyStr=1", replyStr == "1")
}

func do2() {
	addrs := []string{
		"192.168.1.10:7001",
		"192.168.1.10:7002",
		"192.168.1.10:7003",
		"192.168.1.10:7004",
		"192.168.1.10:7005",
	}

	f1 := func() (redis.Conn, error) {
		addr := addrs[0]
		fmt.Println("addr:", addr)
		c, err := redis.Dial("tcp", addr)
		if err != nil {
			fmt.Println(err)
		}
		return c, err
	}

	f2 := func() (redis.Conn, error) {
		addr := addrs[1]
		fmt.Println("addr:", addr)
		c, err := redis.Dial("tcp", addr)
		if err != nil {
			fmt.Println(err)
		}
		return c, err
	}

	f3 := func() (redis.Conn, error) {
		addr := addrs[2]
		fmt.Println("addr:", addr)
		c, err := redis.Dial("tcp", addr)
		if err != nil {
			fmt.Println(err)
		}
		return c, err
	}

	f4 := func() (redis.Conn, error) {
		addr := addrs[3]
		fmt.Println("addr:", addr)
		c, err := redis.Dial("tcp", addr)
		if err != nil {
			fmt.Println(err)
		}
		return c, err
	}

	f5 := func() (redis.Conn, error) {
		addr := addrs[4]
		fmt.Println("addr:", addr)
		c, err := redis.Dial("tcp", addr)
		if err != nil {
			fmt.Println(err)
		}
		return c, err
	}

	var fs []func() (redis.Conn, error)
	fs = append(fs, f1, f2, f3, f4, f5)

	var pools []redsync.Pool
	for _, v := range fs {
		pool := &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial:        v,
		}

		pools = append(pools, pool)
	}

	mu := redsync.New(pools).NewMutex("woshishacha",
		redsync.SetExpiry(30*time.Second),
		redsync.SetTries(10),
		redsync.SetRetryDelay(1*time.Second))

	if err := mu.Lock(); err != nil {
		fmt.Println("AAAAAAAAAAAAAA", err)
	}

	fmt.Println("do something1")

	if err := mu.Lock(); err != nil {
		fmt.Println("BBBBBBBBBBBBBB", err)
	}

	fmt.Println("do something2")

	if ok := mu.Unlock(); !ok {
		fmt.Println("do something3")
	}

	fmt.Println("do something4")
}

func main() {
	do2()
}
