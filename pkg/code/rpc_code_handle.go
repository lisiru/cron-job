package code

import (
	pb "delay-queue/proto"
	"github.com/marmotedu/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGRPCError(err error) error {
	coder:=errors.ParseCoder(err)
	s, _ :=status.New(ToRPCCode(coder.Code()),coder.String()).WithDetails(&pb.Error{Code: int32(coder.Code()),Message: coder.String()})
	return s.Err()
}


func ToRPCCode(code int) codes.Code {
	var statusCode codes.Code
	switch code {
	case ErrRedis:
		statusCode = codes.Internal
	default:
		statusCode = codes.Unknown
	}

	return statusCode
}

type Status struct {
	*status.Status
}

func FromError(err error) *Status  {
	s,_:=status.FromError(err)
	return &Status{s}
}

func ToGRPCStatus(code int,msg string) *Status  {
	s,_:=status.New(ToRPCCode(code),msg).WithDetails(&pb.Error{Code: int32(code), Message: msg})
	return &Status{s}
}
