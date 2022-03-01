package job

import (
	"context"
	"delay-queue/common"
	"delay-queue/params"
	pb "delay-queue/proto"
	"delay-queue/redis"
	"delay-queue/service"
)

type JobHandler struct {
	srv service.Service
	
}

func NewJobHandler(client *redis.RedisClient) *JobHandler  {

	return &JobHandler{srv: service.NewService(client)}

}



// 添加任务
func (j *JobHandler) Add(ctx context.Context,r *pb.AddRequest) (*pb.AddReply,error)  {
	// todo 参数验证 和错误码、错位信息的返回处理
	addJobOptions:=params.AddOrUpdateJobOptions{
		JobId: r.JobId,
		DelaySeconds: r.DelaySeconds,
		TtrSeconds: r.TtrSeconds,
		Body: r.Body,
		IsLoop: r.IsLoop,
		NotifyUrl: r.NotifyUrl,
		Stat: common.JOB_STAT_READY,
	}
	j.srv.JobSrv().AddJob(ctx,addJobOptions)



	return &pb.AddReply{Id: "223333"},nil

	
}

func (j *JobHandler) Del(ctx context.Context,r *pb.DelRequest) (*pb.DelReply,error)  {
	return &pb.DelReply{},nil
}

func (j *JobHandler) Update(ctx context.Context,r *pb.UpdateRequest) (*pb.UpdateReply,error)  {
	return &pb.UpdateReply{Id: "2222"},nil
}






