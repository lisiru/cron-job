package delay_queue

import (
	"delay-queue/config"
	"delay-queue/options"
	"delay-queue/pkg/app"
	"delay-queue/pkg/logger"
	server2 "delay-queue/server"
	"os"
)

func NewApp() error  {
	opts:=options.NewOptions()
	if err:=app.AddConfigToOptions(opts);err!=nil {
		os.Exit(1)
	}

	logger.Init(opts.Log)
	defer logger.Flush()
	cfg,err:=config.CreateConfigFromOptions(opts)
	if err!=nil {
		return err
	}
	stopCh:=server2.SetupSignalHandler()
	return Run(cfg,stopCh)
}