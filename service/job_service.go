package service

import (
	"context"
	"delay-queue/common"
	"delay-queue/params"
	"delay-queue/pkg/code"
	"delay-queue/pkg/logger"
	"delay-queue/pkg/utils"
	"delay-queue/redis"
	"fmt"
	redisV "github.com/go-redis/redis"
	"github.com/marmotedu/errors"
	"reflect"
	"time"
)

type JobSrv interface {
	AddJob(ctx context.Context, opts params.AddOrUpdateJobOptions) (string, error)
	DelJob(ctx context.Context, opts params.DelJobOptions) (string, error)
	UpdateJob(ctx context.Context, opts params.AddOrUpdateJobOptions) (string, error)
	FinishJob(ctx context.Context, opts params.FinishOptions) (string, error)
	TestErr(ctx context.Context) error
	ScanDelayBucket()
	ConsumeReadyJobQueue()
}

type jobService struct {
	client *redis.RedisClient
}

func NewJobService(srv *service) *jobService {
	return &jobService{
		client: srv.client,
	}
}

func (j *jobService) AddJob(ctx context.Context, opts params.AddOrUpdateJobOptions) (string, error) {
	//1、 加入job pool 数据结构为hash key 为jobId为维度，加入前先判断是否已经存在
	// 2、根据delay 算出执行的绝对时间戳，保存到zset中
	jobHashKey := getJobCacheKey(opts.JobId, common.JOB_INFO_HASH_KEY_PREFIX)
	if jobExist, _ := j.JobIsExist(ctx, jobHashKey, opts.JobId); jobExist {
		// 当前任务id已存在，不能添加，请检查任务id
		return "", errors.WithCode(code.ErrJobExist, "")
	}
	// 添加进job pool
	jobData := convertJobOptionToCacheData(opts)

	// 设置delay bucket zset 的数据
	delayBucketJobData := redisV.Z{
		Score:  float64(opts.DelaySeconds + time.Now().Unix()),
		Member: opts.JobId,
	}
	pipeline := j.client.Pipeline()
	pipeline.ZAdd(common.DELAY_JOB_BUKET_ZSET_KEY, delayBucketJobData)
	pipeline.HMSet(jobHashKey, jobData)
	_, err := pipeline.Exec()
	if err != nil {
		return "", errors.WithCode(code.ErrRedis, err.Error())
	}
	return opts.JobId, nil

}

func (j *jobService) DelJob(ctx context.Context, opts params.DelJobOptions) (string, error) {
	// 删除操作，将job从delay bucket 中删除，将job pool 中的状态设置为删除状态
	jobHashKey := getJobCacheKey(opts.JobId, common.JOB_INFO_HASH_KEY_PREFIX)
	if keyExist, _ := j.JobIsExist(ctx, jobHashKey, opts.JobId); !keyExist {
		// 当前任务id不存在，直接返回
		return "", errors.WithCode(code.ErrJobExist, "")
	}
	jobStat := utils.IntToString(common.JOB_STAT_DELETED)
	pipline := j.client.Pipeline()
	pipline.HSet(jobHashKey, opts.JobId, jobStat)
	pipline.ZRem(common.DELAY_JOB_BUKET_ZSET_KEY, opts.JobId)
	_, err := pipline.Exec()
	if err != nil {
		return "", errors.WithCode(code.ErrRedis, err.Error())
	}
	return opts.JobId, nil

}

func (j *jobService) UpdateJob(ctx context.Context, opts params.AddOrUpdateJobOptions) (string, error) {
	//jobHashKey := getJobCacheKey(opts.JobId, common.JOB_INFO_HASH_KEY_PREFIX)
	//// 从hash中取出job的元信息
	//jobData:=j.client.HM
	//
	//
	return "", nil
}

func getJobCacheKey(jobId string, keyPrefix string) string {
	return fmt.Sprintf(keyPrefix, jobId)
}

func (j *jobService) JobIsExist(ctx context.Context, hashKey string, field string) (bool, error) {
	return j.client.HExists(ctx, hashKey, field)
}

// 将添加任务参数转化为map 保存到redis 的hash中作为任务的元信息
func convertJobOptionToCacheData(opts params.AddOrUpdateJobOptions) map[string]interface{} {
	jobDataMap := make(map[string]interface{})
	elem := reflect.ValueOf(&opts).Elem()
	relType := elem.Type()
	for i := 0; i < relType.NumField(); i++ {
		jobDataMap[relType.Field(i).Name] = elem.Field(i).Interface()
	}
	return jobDataMap

}

func (j *jobService) ScanDelayBucket() {
	logger.Info("扫描delaybucket")
	// 加分布式锁
	redisLock := redis.NewRedisLock(redis.WithClient(j.client), redis.WithId(utils.RandomStr(common.RANDOM_STR_LEN)), redis.WithKey(common.JOB_LOCK_KEY))
	// 加锁
	lockRes, _ := redisLock.Lock()
	defer redisLock.ReleaseLock()
	if lockRes {

		// 扫描delay bucket 将score 小于等于当前时间的取出放进ready queue 并设置一个2s的时间差
		var opt = redisV.ZRangeBy{
			Max: utils.Int64ToString(time.Now().Unix() - 2),
		}
		delayJob, err := j.client.ZRangeByScore(common.DELAY_JOB_BUKET_ZSET_KEY, opt)
		if err != nil {

		}
		if len(delayJob) == 0 {
			return
		}
		for _, v := range delayJob {
			// 从hash 保存的job的元信息中获取当前任务的状态，判断是否已经删除
			stat, _ := j.client.HGet(getJobCacheKey(v, common.JOB_INFO_HASH_KEY_PREFIX), "Stat")
			jobStat := utils.StringToInt(stat)
			if jobStat != common.JOB_STAT_DELAY {
				logger.Infof("任务:%s,已经删除状态，不需要执行", v)

			}
			// 正常状态，丢进ready pool
			values, err := j.client.Lpush(common.JOB_READY_QUEUE_KEY, v)
			logger.Infof("data:%s",values)
			if err != nil {
				continue
			}
			// 推进ready pool 后，将该任务从delay bucket 中删除 ,并将任务元信息设置为ready
			pipline := j.client.Pipeline()
			pipline.HSet(getJobCacheKey(v, common.JOB_INFO_HASH_KEY_PREFIX), "Stat", common.JOB_STAT_READY)
			pipline.ZRem(common.DELAY_JOB_BUKET_ZSET_KEY, v)
			_, err = pipline.Exec()
			if err != nil {
				logger.Infof("将任务: 【%s】从delay bucket中移除失败", v)

			}

		}
	}

}

var brPopTimeout = 1 * time.Second

// 从队列中消费任务
func (j *jobService) ConsumeReadyJobQueue() {
	logger.Info("消费job")
	jobId, _ := j.client.BRPop(brPopTimeout, common.JOB_READY_QUEUE_KEY)
	if jobId == nil {
		return
	}
	// 从hash中获取任务的信息，判断任务的状态是否是ready
	jobHashCacheKey := getJobCacheKey(jobId[1], common.JOB_INFO_HASH_KEY_PREFIX)
	jobInfo, _ := j.client.HGetAll(jobHashCacheKey)
	if len(jobInfo)== 0 {
		logger.Errorf("当前任务:【%s】信息不存在,请检查", jobId)
		return
	}
	jobStat := utils.StringToInt(jobInfo["Stat"])
	if jobStat == common.JOB_STAT_DELETED {
		return
	}
	if jobStat == common.JOB_STAT_READY {
		// 调任务方的notify url 通知做任务
		go notifyUrl(jobInfo["NotifyUrl"], jobInfo["body"])
		// 判断当前任务是否为循环任务 和任务执行TTR
		if utils.StringToInt(jobInfo["IsLoop"]) == 1 {
			// 根据ttr 计算下次执行的时间，并设置为delay 状态 ，并加入delay bucket
			nextExecTime := time.Now().Unix() + utils.StringToInt64(jobInfo["TTR"])
			delayBuckeyJobData := redisV.Z{
				Score:  float64(nextExecTime),
				Member: jobId,
			}
			// 重新保存到zset 和更新hash信息中的状态
			_, _ = j.client.Zadd(common.DELAY_JOB_BUKET_ZSET_KEY, delayBuckeyJobData)
			_, _ = j.client.HSet(jobHashCacheKey, "Stat", common.JOB_STAT_DELAY)

			//var redisScript=`redis.call("ZADD",KEYS[1],ARGV[1],ARGV[2]);redis.call("HSET",KEYS[2],ARGV[3],ARGV[4]);`
			//j.client.Eval(redisScript,[]string{common.DELAY_JOB_BUKET_ZSET_KEY,jobHashCacheKey},[]interface{}{nextExecTime,jobId,"Stat",common.JOB_STAT_DELAY})
		} else {
			// 将hash中 的任务状态更新为已取出
			_, _ = j.client.HSet(jobHashCacheKey, "Stat", common.JOB_STAT_RESERVED)
		}

	}

}

func notifyUrl(url string, body string) {

}

// 任务方处理完逻辑后
func (j *jobService) FinishJob(ctx context.Context, opts params.FinishOptions) (string, error) {
	// 判断该jobId 是否存在
	jobHashKey := getJobCacheKey(opts.JobId, common.JOB_INFO_HASH_KEY_PREFIX)
	if jobExist, _ := j.JobIsExist(ctx, jobHashKey, opts.JobId); jobExist {
		// 当前任务id已存在，不能添加，请检查任务id
		return "", errors.WithCode(code.ErrJobExist, "")
	}
	// 判断是否为循环任务
	isLoop, err := j.client.HGet(jobHashKey, "IsLoop")
	if err != nil {
		return "", errors.WithCode(code.ErrJobExist, err.Error())
	}
	if utils.StringToInt(isLoop) == 1 {
		// 循环任务不需要完成
		return "", nil
	}
	// 删除任务信息
	_, err = j.client.HDel(jobHashKey)
	if err != nil {
		return "", errors.WithCode(code.ErrJobExist, err.Error())
	}
	return opts.JobId, nil

}

func (j *jobService) TestErr(ctx context.Context) error {

	return errors.WithCode(code.ErrRedis, "")
}
