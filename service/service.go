package service

import "delay-queue/redis"

type Service interface {
	JobSrv() JobSrv
}

type service struct {
	client *redis.RedisClient
}

func NewService(client *redis.RedisClient) Service  {
	return &service{client: client}
}




func (s *service) JobSrv() JobSrv  {
	return NewJobService(s)
}
