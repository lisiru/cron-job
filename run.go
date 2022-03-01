package delay_queue

import "delay-queue/config"

func Run(cfg *config.Config,stopCh <-chan struct{}) error  {
	server,err:=createServer(cfg)
	if err != nil {
		return err
	}
	return server.PrepareRun().Run(stopCh)
}
