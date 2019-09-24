package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"reflect"
	"time"
)

func main() {
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
