package redis

import (
	"context"
	"crypto/tls"
	"delay-queue/pkg/logger"
	genericoptions "delay-queue/pkg/options"
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"sync"
	"time"
)

type RedisClient struct {
	Client redis.UniversalClient
}

var  (
	redisClient *RedisClient
	once sync.Once
)

func NewRedisClient(opts *genericoptions.RedisOptions) (*RedisClient,error)  {
	logger.Debug("creating new Redis connection pool")
	if opts ==nil && redisClient ==nil{
		return nil,fmt.Errorf("failed to new redis cache factory")
	}
	var client redis.UniversalClient
	once.Do(func() {
		poolSize:=500
		if opts.MaxActive>0 {
			poolSize=opts.MaxActive
		}

		timeout:=5*time.Second
		if opts.Timeout>0 {
			timeout=time.Duration(opts.Timeout)*time.Second
		}
		var tlsConfig *tls.Config
		if opts.UseSSL{
			tlsConfig=&tls.Config{
				InsecureSkipVerify: opts.SSLInsecureSkipVerify,
			}
		}
		options:=&RedisOption{
			Addrs: getRedisAddrs(opts),
			MasterName: opts.MasterName,
			Password: opts.Password,
			DB: opts.Database,
			DialTimeout: timeout,
			ReadTimeout: timeout,
			WriteTimeout: timeout,
			IdleTimeout: 240*timeout,
			PoolSize: poolSize,
			TLSConfig: tlsConfig,
		}
		logger.Info("--> [REDIS] create single-node")
		client = redis.NewClient(options.simple())
		redisClient = &RedisClient{Client: client}
	})
	return redisClient,nil
}
type RedisOption redis.UniversalOptions

func getRedisAddrs(opts *genericoptions.RedisOptions) (addrs []string)  {
	if len(opts.Addrs) !=0 {
		addrs = opts.Addrs
	}
	if len(addrs) ==0 &&opts.Port!=0 {
		addr:=opts.Host+":" + strconv.Itoa(opts.Port)
		addrs=append(addrs,addr)
	}
	return addrs
}
func (o *RedisOption) simple() *redis.Options  {

	addr:="127.0.0.1:6379"
	if len(o.Addrs)>0{
		addr = o.Addrs[0]
	}
	return &redis.Options{
		Addr: addr,
		OnConnect: o.OnConnect,
		DB: o.DB,
		Password: o.Password,
		MaxRetries: o.MaxRetries,
		MinRetryBackoff: o.MinRetryBackoff,
		MaxRetryBackoff: o.MaxRetryBackoff,
		DialTimeout: o.DialTimeout,
		ReadTimeout: o.ReadTimeout,
		WriteTimeout: o.WriteTimeout,
		PoolSize: o.PoolSize,
		MinIdleConns: o.MinIdleConns,
		MaxConnAge: o.MaxConnAge,
		PoolTimeout: o.PoolTimeout,
		IdleTimeout: o.IdleTimeout,
		IdleCheckFrequency: o.IdleCheckFrequency,
		TLSConfig: o.TLSConfig,
	}
}

func (r *RedisClient) Zadd(key string,members redis.Z) (int64,error)  {
	return r.Client.ZAdd(key,members).Result()
}

func (r *RedisClient) ZRangeByScore(key string,opt redis.ZRangeBy) ([]string,error) {
	return r.Client.ZRangeByScore(key,opt).Result()
}

func (r *RedisClient) Lpush(key string,values interface{}) (int64,error)  {
	return r.Client.LPush(key,values).Result()
}

func (r *RedisClient) BRPop(timeout time.Duration,keys ...string) ([]string,error) {
	return r.Client.BRPop(timeout,keys...).Result()
}

func (r *RedisClient) HExists(ctx context.Context,key string,field string) (bool,error) {
	return r.Client.HExists(key,field).Result()
}

func (r *RedisClient) HGet(key string,field string) (string,error)  {
	return r.Client.HGet(key,field).Result()
}

func (r *RedisClient) HGetAll(key string) (map[string]string,error) {
	return r.Client.HGetAll(key).Result()
}

func (r *RedisClient) HDel(key string,field ...string) (int64,error)   {
	return r.Client.HDel(key,field...).Result()
}

func (r *RedisClient) HSet(key string,filed string,values interface{}) (bool,error) {
	return r.Client.HSet(key,filed,values).Result()
}

func (r *RedisClient) HMSet(key string,fields map[string]interface{}) (string,error) {
	return r.Client.HMSet(key,fields).Result()
}

func (r *RedisClient) HMget(key string,fields ...string) ([]interface{},error) {
	return r.Client.HMGet(key,fields...).Result()
}





func (r *RedisClient) Pipeline() redis.Pipeliner  {
	return r.Client.Pipeline()
}

func (r *RedisClient) ZRem(key string,members ...interface{}) (int64,error) {
	return r.Client.ZRem(key,members).Result()
}

func (r *RedisClient) Eval(script string,keys []string,args ...interface{}) (interface{},error)   {
	return r.Client.Eval(script,keys,args).Result()
}






