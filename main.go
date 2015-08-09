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

	lis, err := NewLis(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	errorChan := make(chan error)

	// run lis
	go lis.run(errorChan)

	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case sig := <-signalChan:
			switch sig {
			case os.Interrupt:
				fmt.Println("os Interrupt")
			case syscall.SIGTERM:
				fmt.Println("SIGTERM")
				// err := lis.storeState()
				// if err != nil {
				// 	fmt.Fprintln(os.Stderr, err.Error())
				// }
			}
		case err = <-errorChan:
			// TODO better logging
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
}
