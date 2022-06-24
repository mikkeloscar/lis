package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/mikkeloscar/lis"
	log "github.com/sirupsen/logrus"
)

func main() {
	confPath := flag.String("c", "/etc/lis.conf", "Config file")
	flag.Parse()

	config, err := lis.ReadConfig(*confPath)
	if err != nil {
		log.Fatal(err)
	}

	l, err := lis.NewLis(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go handleSigterm(cancel)

	// run lis
	err = l.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func handleSigterm(cancelFunc func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	log.Info("Received SIGTERM. Terminating...")
	cancelFunc()
}
