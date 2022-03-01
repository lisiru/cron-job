package hello

import (
	"context"
	helloworld "delay-queue/proto"
)

type GreeterServer struct {

}

func (s *GreeterServer) SayHello(ctx context.Context, r *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: "hello world"+r.Name},nil

}
