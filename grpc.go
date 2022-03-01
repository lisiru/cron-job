package delay_queue

import (
	"delay-queue/pkg/logger"
	"google.golang.org/grpc"
	"net"
)

type grpcApiServer struct {
	*grpc.Server
	address string
}

func (s *grpcApiServer) Run()  {
	listen,err:=net.Listen("tcp",s.address)
	if err != nil {
		logger.Fatalf("failed to listen: %s",err.Error())
	}
	go func() {
		if err:=s.Serve(listen);err!=nil{
			logger.Fatalf("failed to start grpc server: %s",err.Error())

		}
	}()

	logger.Infof("start grpc server at %s",s.address)

}

func (s *grpcApiServer) Close()  {
	s.GracefulStop()
	logger.Infof("GRPC server on %s stopped",s.address)
}


