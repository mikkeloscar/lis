package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// Lis defines the core functionality of the lis daemon
type Lis struct {
	current   uint16         // current brightness value
	idleMode  bool           // true if in idle mode
	state     StateFile      // state file
	backlight *Backlight     // backlight
	input     chan struct{}  // input channel used to notify about activity when in idle mode
	idle      chan struct{}  // idle channel used when user is idle
	power     chan struct{}  // power channel used to notify about power changes (AC/Battery)
	signals   chan os.Signal // signals channel used to handle external signals
	errors    chan error     // errors channel
	idleTime  uint           // idle time in minutes
}

// NewLis creates a new Lis instance
func NewLis(config *Config, sigChan chan os.Signal) (*Lis, error) {
	if config.Backlight != "intel" {
		return nil, fmt.Errorf("backlight: %s not supported", config.Backlight)
	}

	backlight, err := NewBacklight("intel_backlight")
	if err != nil {
		return nil, err
	}

	return &Lis{
		idleMode:  false,
		state:     StateFile(config.StateFile),
		backlight: backlight,
		input:     make(chan struct{}),
		idle:      make(chan struct{}),
		power:     make(chan struct{}),
		signals:   sigChan,
		errors:    make(chan error),
		idleTime:  config.IdleTime,
	}, nil
}

// load state from stateFile
func (l *Lis) loadState() error {
	v, err := l.state.Read()
	if err != nil {
		return err
	}

	l.current = v

	return nil
}

// store current state in stateFile
func (l *Lis) storeState() error {
	return l.state.Write(l.current)
}

// get current brightness level
func (l *Lis) getCurrent() error {
	v, err := l.backlight.Get()
	if err != nil {
		return err
	}

	l.current = uint16(v)

	return nil
}

// lis main loop
func (l *Lis) run() {
	var err error

	// start Listening for idle
	l.idleListener()

	for {
		select {
		case <-l.input:
			// undim screen
			l.unDim()
			l.idleMode = false

			// start Listening for idle
			l.idleListener()
		case <-l.idle:
			fmt.Println("enter idle mode")
			// get current brightness level
			err = l.getCurrent()
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				continue
			}

			// dim screen
			l.dim()
			l.idleMode = true

			// start Listening for input to exit idle mode
			err = l.inputListener()
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				continue
			}
		case power := <-l.power:
			fmt.Println("power", power)
		case sig := <-l.signals:
			switch sig {
			case syscall.SIGTERM:
				exit := 0

				if l.idleMode {
					err = l.storeState()
					if err != nil {
						fmt.Fprintln(os.Stderr, err.Error())
						exit = 1
					}
				} else {
					err = l.getCurrent()
					if err != nil {
						fmt.Fprintln(os.Stderr, err.Error())
						exit = 1
					}

					err = l.storeState()
					if err != nil {
						fmt.Fprintln(os.Stderr, err.Error())
						exit = 1
					}
				}

				os.Exit(exit)
			}
		case err := <-l.errors:
			// Write error to stderr
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
}

// dim screen
func (l *Lis) dim() {
	fmt.Printf("dimming screen from brightness level: %d\n", l.current)
	go l.backlight.Dim(int(l.current), 0, l.errors)
}

// undim screen
func (l *Lis) unDim() {
	fmt.Printf("undimming screen to brightness level: %d\n", l.current)
	go l.backlight.UnDim(0, int(l.current), l.errors)
}

// listen for input activity
func (l *Lis) inputListener() error {
	devices, err := GetInputDevices(l.errors)
	if err != nil {
		return err
	}

	go devices.Wait(l.input)

	return nil
}

// listen for user idling
func (l *Lis) idleListener() {
	go l.xidle()
}

// listen for X idletime
func (l *Lis) xidle() {
	for {
		time.Sleep(time.Duration(l.idleTime/3) * time.Millisecond)
		idleTime, err := XIdle()
		if err != nil {
			l.errors <- err
			continue
		}

		fmt.Println(idleTime)

		if idleTime >= l.idleTime {
			l.idle <- struct{}{}
			break
		}
	}
}
