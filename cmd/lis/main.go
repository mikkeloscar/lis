package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mikkeloscar/lis"
)

func main() {
	confPath := flag.String("c", "/etc/lis.conf", "Config file")
	flag.Parse()

	config, err := lis.ReadConfig(*confPath)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	l, err := lis.NewLis(config)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go handleSigterm(cancel)

	// run lis
	err = l.Run(ctx)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func handleSigterm(cancelFunc func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	slog.Info("Received SIGTERM. Terminating...")
	cancelFunc()
}
