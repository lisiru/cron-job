package job

import (
	"context"
	"delay-queue/common"
	"delay-queue/params"
	"delay-queue/pkg/code"
	"delay-queue/pkg/logger"
	pb "delay-queue/proto"
	"delay-queue/redis"
	"delay-queue/service"
	"runtime"
	"time"
)

type JobHandler struct {
	srv service.Service
	scanSignal chan bool
}

func NewJobHandler(client *redis.RedisClient) *JobHandler {

	return &JobHandler{srv: service.NewService(client)}

}

// 添加任务
func (j *JobHandler) Add(ctx context.Context, r *pb.AddRequest) (*pb.AddReply, error) {
	// todo 参数验证 和错误码、错位信息的返回处理
	addJobOptions := params.AddOrUpdateJobOptions{
		JobId:        r.JobId,
		DelaySeconds: r.DelaySeconds,
		TtrSeconds:   r.TtrSeconds,
		Body:         r.Body,
		IsLoop:       r.IsLoop,
		NotifyUrl:    r.NotifyUrl,
		Stat:         common.JOB_STAT_DELAY,
	}
	res, err := j.srv.JobSrv().AddJob(ctx, addJobOptions)
	if err != nil {
		return nil, code.ToGRPCError(err)
	}
	return &pb.AddReply{JobId: res}, nil

}

func (j *JobHandler) Del(ctx context.Context, r *pb.DelRequest) (*pb.DelReply, error) {
	delOption := params.DelJobOptions{JobId: r.JobId}
	_, err := j.srv.JobSrv().DelJob(ctx, delOption)
	if err != nil {
		return nil, code.ToGRPCError(err)
	}
	return &pb.DelReply{}, nil
}

func (j *JobHandler) Update(ctx context.Context, r *pb.UpdateRequest) (*pb.UpdateReply, error) {
	return &pb.UpdateReply{JobId: "2222"}, nil
}

func (j *JobHandler) Finish(ctx context.Context, r *pb.FinishRequest) (*pb.FinishReply, error) {

	finishOption := params.FinishOptions{JobId: r.JobId}
	_, err := j.srv.JobSrv().FinishJob(ctx, finishOption)
	if err != nil {
		return nil, code.ToGRPCError(err)
	}

	return &pb.FinishReply{}, nil
}

func (j *JobHandler) TestErrCode(ctx context.Context, r *pb.TestErrRequest) (*pb.TestErrReply, error) {
	err := j.srv.JobSrv().TestErr(ctx)
	if err != nil {
		return nil, code.ToGRPCError(err)
	}
	return nil, nil
}

func (j *JobHandler) ScanJob()  {
	defer func() {
		logger.Info("任务池扫描结束")
	}()

	tick:=time.NewTicker(time.Second)

	for  {
		select {
		case <-j.scanSignal:
			logger.Info("收到停止信号，扫描任务终止")
			tick.Stop()
			return
		case <-tick.C:
			go j.srv.JobSrv().ScanDelayBucket()
			go j.srv.JobSrv().ConsumeReadyJobQueue()

		}
		logger.Infof("goroutines: %d", runtime.NumGoroutine())
	}
}
