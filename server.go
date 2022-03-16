package delay_queue

import (
	"delay-queue/config"
	genericoptions "delay-queue/pkg/options"
	pb "delay-queue/proto"
	"delay-queue/server/job"
	"net/http"

	//helloworld "delay-queue/proto"
	"delay-queue/redis"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	_ "net/http/pprof"
)

type server struct {
	grpcApiServer *grpcApiServer
	jobHandler *job.JobHandler

}

type preparedAPIServer struct {
	*server
}

type GenericConfig struct {
	Addr         string
	MaxMsgSize   int
	redisOptions *genericoptions.RedisOptions
}

func createServer(cfg *config.Config) (*server,error)  {
	genericConfig,err:= buildGenericConfig(cfg)
	if err != nil {
		return nil, err
	}

	grpcServer,err:=genericConfig.New()

	server:=&server{
		grpcApiServer:grpcServer,
	}
	return server,nil

}

func (s *server) PrepareRun() preparedAPIServer  {
	// 可以做一些前置准备工作 ，比如初始化路由，初始化redis等
	//s.initRedis()
	return preparedAPIServer{s}
}
// 初始化redis
//func (s *server) initRedis()  {
//	go func() {
//		_, _ = redis.NewRedisClient(s.redisOptions)
//	}()
//}

func (s preparedAPIServer) Run(stopCh <-chan struct{}) error  {
	go s.grpcApiServer.Run()
	go func() {
		_ = http.ListenAndServe(":6060", nil)
	}()
	<-stopCh
	s.grpcApiServer.Close()
	return nil
}

func (c *GenericConfig) New() (*grpcApiServer,error)  {
	client, _ := redis.NewRedisClient(c.redisOptions)
	opts:=[]grpc.ServerOption{grpc.MaxRecvMsgSize(c.MaxMsgSize)}
	grpcServer :=grpc.NewServer(opts...)
	jobHandler:=job.NewJobHandler(client)
	pb.RegisterJobServer(grpcServer,jobHandler)
	//helloworld.RegisterGreeterServer(grpcServer, &hello.GreeterServer{})
	//helloworld.RegisterGreeterServer(grpcServer,&Greeter)
	//pb.RegisterCacheServer(grpcServer, cacheIns)

	reflection.Register(grpcServer)
	go jobHandler.ScanJob()

	return &grpcApiServer{grpcServer, c.Addr}, nil

}

func buildGenericConfig(cfg *config.Config) (*GenericConfig, error) {
	return &GenericConfig{
		Addr:         fmt.Sprintf("%s:%d", cfg.GRPCOptions.BindAddress, cfg.GRPCOptions.BindPort),
		MaxMsgSize:   cfg.GRPCOptions.MaxMsgSize,
		redisOptions: cfg.RedisOptions,
	}, nil
}
