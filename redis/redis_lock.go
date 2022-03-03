package redis

import (
	"delay-queue/common"
	"delay-queue/pkg/logger"
	"delay-queue/pkg/utils"
	"github.com/go-redis/redis"
	"github.com/vmihailenco/msgpack"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type RedisLock struct {
	client *RedisClient
	expire uint32
	key    string
	id     string
}

func (r RedisLock) MarshalBinary() ([]byte,error) {
	return msgpack.Marshal(r)
}
func (r RedisLock) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, r)
}

var mux sync.Mutex
var (
	defaultExpire   uint32 = 1
	millisPerSecond        = 1000
)

type Option func(*RedisLock)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewRedisLock(opts ...func(*RedisLock)) *RedisLock {
	redisLock := &RedisLock{
		expire: defaultExpire,
	}
	for _, option := range opts {
		option(redisLock)
	}
	return redisLock
}

func WithClient(client *RedisClient) Option {

	return func(lock *RedisLock) {
		mux.Lock()
		defer mux.Unlock()
		lock.client = client
	}

}

func WithExpire(expire uint32) Option {
	return func(lock *RedisLock) {
		mux.Lock()
		defer mux.Unlock()
		atomic.StoreUint32(&lock.expire,expire)
	}

}

func WithKey(key string) Option {
	return func(lock *RedisLock) {
		mux.Lock()
		defer mux.Unlock()
		lock.key = key
	}
}

func WithId(id string) Option {
	return func(lock *RedisLock) {
		mux.Lock()
		defer mux.Unlock()
		lock.id = id
	}
}

func (l *RedisLock) Lock() (bool, error) {
	expires := atomic.LoadUint32(&l.expire)
	resp, err := l.client.Eval(common.LOCK_COMMAND, []string{l.key}, []string{l.id, utils.IntToString(int(expires) * millisPerSecond)})
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		logger.Errorf("error on locking for %s %s", l.key, err.Error())
		return false, err
	} else if resp == nil {
		return false, nil
	}
	reply, ok := resp.(string)
	if ok && reply == "OK" {
		//logger.Info("加锁成功")
		return true, nil

	}
	logger.Errorf("Unknown reply when acquiring lock for %s: %v", l.key, resp)
	return false, nil

}

func (l *RedisLock) ReleaseLock() (bool,error) {
	resp,err:=l.client.Eval(common.DEL_COMMAND,[]string{l.key},[]string{l.id})
	if err != nil {
		//logger.Info("解锁错误")
		return false, err
	}
	reply,ok:=resp.(int64)
	if !ok{
		//logger.Info("解锁失败")
		return false,nil
	}
	//logger.Info("解锁成功")
	return reply==1,nil
}
