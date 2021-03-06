package redis

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/harveywangdao/ants/logger"
	"github.com/harveywangdao/ants/util"
	"time"
)

type Redis struct {
	conn redis.Conn
}

type RedisPool struct {
	pool *redis.Pool
}

const (
	MAX_POOL_SIZE  = 20
	MAX_IDLE_NUM   = 2
	MAX_ACTIVE_NUM = 20
	REDIS_ADDR     = "localhost:6379"
	REDISPASSWORD  = "180498"
)

func NewRedisPool(addr, pw string) (*RedisPool, error) {
	pool := &redis.Pool{
		MaxIdle:     MAX_IDLE_NUM,
		MaxActive:   MAX_ACTIVE_NUM,
		IdleTimeout: (60 * time.Second),
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				logger.Error(err)
				return nil, err
			}

			/*if _, err := c.Do("AUTH", pw); err != nil {
				c.Close()
				return nil, err
			}*/
			/*
			   if _, err := c.Do("SELECT", db); err != nil {
			     c.Close()
			     return nil, err
			   }*/
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	return &RedisPool{
		pool: pool,
	}, nil
}

func (rp *RedisPool) Get() (*Redis, error) {
	red := &Redis{}
	red.conn = rp.pool.Get()
	return red, nil
}

func NewRedis(pool *redis.Pool) (*Redis, error) {
	red := &Redis{}
	red.conn = pool.Get()
	return red, nil
}

type DistLock struct {
	pool    *RedisPool
	key     string
	value   string
	timeout int64
	conn    *Redis
}

func NewDistLock(pool *RedisPool, key string, timeout int64) *DistLock {
	return &DistLock{
		pool:    pool,
		key:     key,
		value:   util.GetUUID(),
		timeout: timeout,
	}
}

func (l *DistLock) Lock() error {
	c, err := l.pool.Get()
	if err != nil {
		logger.Error(err)
		return err
	}
	defer c.Close()

	if err := c.Lock(l.key, l.value, l.timeout); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (l *DistLock) Unlock() {
	c, err := l.pool.Get()
	if err != nil {
		logger.Error(err)
		return
	}
	defer c.Close()

	if err := c.Unlock(l.key, l.value); err != nil {
		logger.Error(err)
	}
}

func (l *DistLock) Lock1() error {
	if l.conn == nil {
		c, err := l.pool.Get()
		if err != nil {
			logger.Error(err)
			return err
		}
		l.conn = c
	}

	if err := l.conn.Lock(l.key, l.value, l.timeout); err != nil {
		logger.Error(l.key, "redis lock fail:", err)
		l.conn.Close()
		l.conn = nil
		return err
	}

	return nil
}

func (l *DistLock) Unlock1() {
	if l.conn == nil {
		c, err := l.pool.Get()
		if err != nil {
			logger.Error(err)
			return
		}
		l.conn = c
	}

	defer func() {
		l.conn.Close()
		l.conn = nil
	}()

	if err := l.conn.Unlock(l.key, l.value); err != nil {
		logger.Error(l.key, "redis unlock fail:", err)
	}
}

func (red *Redis) Close() error {
	return red.conn.Close()
}

func (red *Redis) Err() error {
	return red.conn.Err()
}

func (red *Redis) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	return red.conn.Do(commandName, args...)
}

func (red *Redis) Send(commandName string, args ...interface{}) error {
	return red.conn.Send(commandName, args...)
}

func (red *Redis) Flush() error {
	return red.conn.Flush()
}

func (red *Redis) Receive() (reply interface{}, err error) {
	return red.conn.Receive()
}

/*
锁失败：
1.已上锁
2.网络问题
*/
func (red *Redis) Lock(key, value string, timeout int64) error {
	reply, err := red.conn.Do("SET", key, value, "NX", "EX", timeout)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("Lock", key, value)

	setOk, ok := reply.(string)
	if ok && setOk == "OK" {
		return nil
	}

	return errors.New(fmt.Sprintf("Lock fail, key %s already existed", key))
}

func (red *Redis) Unlock(key, value string) error {
	sc := redis.NewScript(1, `
if redis.call("get",KEYS[1]) == ARGV[1]
then
    return redis.call("del",KEYS[1])
else
    return 0
end`)

	reply, err := sc.Do(red.conn, key, value)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("Unlock", key, value)

	if fmt.Sprintf("%v", reply) == "1" {
		return nil
	}

	return fmt.Errorf("Unlock fail, key %s already deleted", key)
}

func (red *Redis) IsKeyExist(key string) bool {
	exists, _ := redis.Bool(red.conn.Do("EXISTS", key))
	return exists
}

func (red *Redis) DeleteKey(key string) error {
	n, err := red.conn.Do("DEL", key)
	if err != nil {
		logger.Error(err, n)
		return err
	}

	return nil
}

func (red *Redis) CreateListByInt64Slice(key string, data []int64) error {
	//judge key exist
	if red.IsKeyExist(key) {
		logger.Warn(key, "Existed.")
		err := red.DeleteKey(key)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	//insert data
	for i := 0; i < len(data); i++ {
		_, err := red.conn.Do("RPUSH", key, data[i])
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	return nil
}

func (red *Redis) GetListValueByIndex(key string, index int) (int64, error) {
	v, err := redis.Int64(red.conn.Do("LINDEX", key, index))
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	return v, nil
}

func (red *Redis) GetListLen(key string) (int, error) {
	v, err := redis.Int(red.conn.Do("LLEN", key))
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	return v, nil
}

func (red *Redis) GetInt64SliceList(key string) ([]int64, error) {
	l, err := red.GetListLen(key)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	v, err := redis.Int64s(red.conn.Do("LRANGE", key, 0, l))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return v, nil
}

func (red *Redis) SetListValueByIndex(key string, index int, value int64) error {
	_, err := red.conn.Do("LSET", key, index, value)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (red *Redis) Hset(key, field, value string) error {
	_, err := red.conn.Do("HSET", key, field, value)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (red *Redis) Hget(key, field string) (string, error) {
	value, err := redis.String(red.conn.Do("HGET", key, field))
	if err != nil {
		logger.Error(err)
		return "", err
	}

	return value, nil
}

func (red *Redis) Hmset(key string, m map[string]interface{}) error {
	var args []interface{}
	args = append(args, key)
	for k, v := range m {
		args = append(args, k)
		args = append(args, v)
	}

	_, err := red.conn.Do("HMSET", args...)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (red *Redis) Hmget(key string, fields []string) ([]string, error) {
	var args []interface{}
	args = append(args, key)
	for _, v := range fields {
		args = append(args, v)
	}

	value, err := redis.Strings(red.conn.Do("HGET", args...))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return value, nil
}

func (red *Redis) Hgetall(key string) (map[string]string, error) {
	fieldValues, err := redis.StringMap(red.conn.Do("HGETALL", key))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return fieldValues, nil
}

func (red *Redis) ListPush(key, value string) error {
	_, err := red.conn.Do("LPUSH", key, value)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (red *Redis) ListPop(key string) (string, error) {
	value, err := redis.String(red.conn.Do("RPOP", key))
	if err == redis.ErrNil {
		logger.Debug("list", key, "is nil")
		return "", nil
	}

	if err != nil {
		logger.Error(err)
		return "", err
	}

	return value, nil
}

func (red *Redis) SetAdd(key, value string) error {
	_, err := red.conn.Do("SADD", key, value)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (red *Redis) SetLen(key string) (int64, error) {
	value, err := redis.Int64(red.conn.Do("SCARD", key))
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	return value, nil
}

func (red *Redis) SetMembers(key string) ([]string, error) {
	value, err := redis.Strings(red.conn.Do("SMEMBERS", key))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return value, nil
}

func (red *Redis) ZsetAdd(key, value string, score int64) error {
	_, err := red.conn.Do("ZADD", key, score, value)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (red *Redis) ZsetLen(key string) (int64, error) {
	value, err := redis.Int64(red.conn.Do("ZCARD", key))
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	return value, nil
}

func (red *Redis) ZsetLenBetweenScores(key string, min, max int64) (int64, error) {
	value, err := redis.Int64(red.conn.Do("ZCOUNT", key, min, max))
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	return value, nil
}

func (red *Redis) ZsetMembers(key string) (map[string]int64, error) {
	//value, err := redis.Int64Map(red.conn.Do("ZRANGE", key, 0, -1, "WITHSCORES"))
	value, err := redis.Int64Map(red.conn.Do("ZREVRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return value, nil
}

func (red *Redis) HyperLogLogAdd(key, value string) error {
	_, err := red.conn.Do("PFADD", key, value)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (red *Redis) HyperLogLogLen(key string) (int64, error) {
	value, err := redis.Int64(red.conn.Do("PFCOUNT", key))
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	return value, nil
}

func (red *Redis) Publish(channel, msg string) error {
	_, err := red.conn.Do("PUBLISH", channel, msg)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}
