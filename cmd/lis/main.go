package main

import (
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

	stopChan := make(chan struct{})

	l, err := lis.NewLis(config, stopChan)
	if err != nil {
		log.Fatal(err)
	}

	go handleSigterm(stopChan)

	// run lis
	err = l.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func handleSigterm(stop chan<- struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	switch sig {
	case syscall.SIGINT, syscall.SIGTERM:
		log.Info("Received SIGTERM. Terminating...")
		close(stop)
	}
}
