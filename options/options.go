package options

import (
	"delay-queue/pkg/logger"
	genericoptions "delay-queue/pkg/options"
)

type Options struct {
	GRPCOptions *genericoptions.GRPCOptions `json:"grpc" mapstructure:"grpc"`
	RedisOptions *genericoptions.RedisOptions `json:"redis" mapstructure:"redis"`
	Log *logger.Options `json:"log" mapstructure:"log"`

}

func NewOptions() *Options  {
	o:=Options{
		GRPCOptions: genericoptions.NewGRPCOptions(),
		RedisOptions: genericoptions.NewRedisOptions(),
		Log: logger.NewOptions(),

	}
	return &o
}
