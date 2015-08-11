package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	confPath := flag.String("c", "/etc/lis.conf", "Config file")
	flag.Parse()

	config, err := ReadConfig(*confPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	signalChan := make(chan os.Signal, 2)
	// signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	signal.Notify(signalChan, syscall.SIGTERM)

	lis, err := NewLis(config, signalChan)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// run lis
	lis.run()
}
