package config

import "delay-queue/options"

type Config struct {
	*options.Options
}

func CreateConfigFromOptions(opts *options.Options) (*Config,error)  {
	return &Config{
		opts,

	},nil
}
