package service

import (
	"context"
	"delay-queue/common"
	"delay-queue/params"
	"delay-queue/pkg/logger"
	"delay-queue/pkg/utils"
	"delay-queue/redis"
	"fmt"
	redisV "github.com/go-redis/redis"
	"reflect"
	"time"
)

type JobSrv interface {
	AddJob(ctx context.Context, opts params.AddOrUpdateJobOptions) string
	DelJob(ctx context.Context, opts params.DelJobOptions) string
	UpdateJob(ctx context.Context, opts params.AddOrUpdateJobOptions) string
}

type jobService struct {
	client *redis.RedisClient
}

func NewJobService(srv *service) *jobService {
	return &jobService{
		client: srv.client,
	}
}

func (j *jobService) AddJob(ctx context.Context, opts params.AddOrUpdateJobOptions) string {
	//1、 加入job pool 数据结构为hash key 为jobId为维度，加入前先判断是否已经存在
	// 2、根据delay 算出执行的绝对时间戳，保存到zset中
	jobHashKey := getJobCacheKey(opts.JobId, common.JOB_INFO_HASH_KEY_PREFIX)
	if jobExist, _ := j.JobIsExist(ctx, jobHashKey, opts.JobId); jobExist {
		// 当前任务id已存在，不能添加，请检查任务id
		return ""
	}
	// 添加进job pool
	jobData := convertJobOptionToCacheData(opts)

	// 设置delay buket 的数据
	delayBucketJobData := redisV.Z{
		Score:  float64(opts.DelaySeconds),
		Member: opts.JobId,
	}
	pipeline := j.client.Pipeline()
	_ = []*redisV.IntCmd{pipeline.ZAdd(common.DELAY_JOB_BUKET_ZSET_KEY, delayBucketJobData)}
	_ = []*redisV.StatusCmd{pipeline.HMSet(jobHashKey, jobData)}
	_, err := pipeline.Exec()
	if err != nil {
		// todo 错误处理
		return ""
	}
	return opts.JobId

}

func (j *jobService) DelJob(ctx context.Context, opts params.DelJobOptions) string {
	// 删除操作，将job从delay bucket 中删除，将job pool 中的状态设置为删除状态
	jobHashKey := getJobCacheKey(opts.JobId, common.JOB_INFO_HASH_KEY_PREFIX)
	if keyExist, _ := j.JobIsExist(ctx, jobHashKey, opts.JobId); !keyExist {
		// 当前任务id不存在，直接返回
		// todo 相关错误状态码设置返回
		return ""
	}
	jobStat := utils.IntToString(common.JOB_STAT_DELETED)
	pipline := j.client.Pipeline()
	pipline.HSet(jobHashKey, opts.JobId, jobStat)
	pipline.ZRem(common.DELAY_JOB_BUKET_ZSET_KEY, opts.JobId)
	_, err := pipline.Exec()
	if err != nil {
		// todo 错误处理
		return ""
	}
	return opts.JobId

}

func (j *jobService) UpdateJob(ctx context.Context, opts params.AddOrUpdateJobOptions) string {
	//jobHashKey := getJobCacheKey(opts.JobId, common.JOB_INFO_HASH_KEY_PREFIX)
	//// 从hash中取出job的元信息
	//jobData:=j.client.HM
	//
	//
	return ""
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

func (j *jobService) ScanDelayBucket()  {
	// 加分布式锁
	redisLock:=redis.NewRedisLock(redis.WithClient(j.client),redis.WithId(utils.RandomStr(common.RANDOM_STR_LEN)),redis.WithKey(common.JOB_LOCK_KEY))
	// 加锁
	lockRes,_:=redisLock.Lock()
	defer redisLock.ReleaseLock()
	if lockRes{
		// 扫描delay bucket 将score 大于等于当前时间的取出放进ready queue 并设置一个2s的时间差
		var opt=redisV.ZRangeBy{
			Max:    utils.Int64ToString(time.Now().Unix()-2),
		}
		delayJob, _ :=j.client.ZRangeByScore(common.DELAY_JOB_BUKET_ZSET_KEY,opt)
		if len(delayJob)==0 {
			return
		}
		for _,v:=range delayJob{
			// 从hash 保存的job的元信息中获取当前任务的状态，判断是否已经删除
			stat,_:=j.client.HGet(getJobCacheKey(v,common.JOB_INFO_HASH_KEY_PREFIX),"stat")
			jobStat:=utils.StringToInt(stat)
			if jobStat != common.JOB_STAT_DELAY {
				logger.Infof("任务:%s,已经删除状态，不需要执行",v)
				continue
			}
			// 正常状态，丢进ready pool
			_, err := j.client.Lpush(common.JOB_READY_QUEUE_KEY, v)
			if err != nil {
				continue
			}
			// 推进ready pool 后，将该任务从delay bucket 中删除 ,并将任务元信息设置为ready
			pipline := j.client.Pipeline()
			pipline.HSet(getJobCacheKey(v,common.JOB_INFO_HASH_KEY_PREFIX),"stat",common.JOB_STAT_READY)
			pipline.ZRem(common.DELAY_JOB_BUKET_ZSET_KEY,v)
			_, err = pipline.Exec()
			if err != nil {
				logger.Infof("将任务: 【%s】从delay bucket中移除失败",v)

			}


		}
	}

}

var brPopTimeout =1*time.Second

// 从队列中消费任务
func (j *jobService) ConsumeReadyJobQueue()  {
	jobId,_:=j.client.BRPop(brPopTimeout,common.JOB_READY_QUEUE_KEY)
	if jobId==nil{
		return
	}
	// 从hash中获取任务的信息，判断任务的状态是否是ready
	jobInfo, _ :=j.client.HGetAll(getJobCacheKey(jobId[0],common.JOB_INFO_HASH_KEY_PREFIX))
	if jobInfo == nil{
		logger.Errorf("当前任务:【%s】信息不存在,请检查",jobId)
		return
	}
	if utils.StringToInt(jobInfo["stat"]) == common.JOB_STAT_READY{


	}

}

func notifyUrl(url string)  {

}
