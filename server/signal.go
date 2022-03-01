package server

import (
	"os"
	"os/signal"
	"syscall"
)

var onlyOneSignalHandler = make(chan struct{})

var shutdownHandler chan os.Signal

var shutdownSignals =[]os.Signal{os.Interrupt,syscall.SIGTERM}

func SetupSignalHandler() <-chan struct{} {
	close(onlyOneSignalHandler) // panics when called twice

	shutdownHandler = make(chan os.Signal, 2)

	stop := make(chan struct{})

	signal.Notify(shutdownHandler, shutdownSignals...)

	go func() {
		<-shutdownHandler
		close(stop)
		<-shutdownHandler
		os.Exit(1)
	}()

	return stop
}
