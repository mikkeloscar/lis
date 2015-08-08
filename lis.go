package main

import (
	"fmt"
	"time"
)

type Lis struct {
	current   uint16        // current brightness value
	state     StateFile     // state file
	backlight *Backlight    // backlight
	input     chan struct{} // input channel used to notify about activity when in idle mode
	idle      chan struct{} // idle channel used when user is idle
	power     chan struct{} // power channel used to notify about power changes (AC/Battery)
	signals   chan struct{} // signals channel used to handle external signals
	errors    chan error    // errors channel
	idleTime  uint          // idle time in minutes
}

func NewLis(statefile string, idleTime uint) (*Lis, error) {
	backlight, err := NewBacklight("intel_backlight")
	if err != nil {
		return nil, err
	}

	return &Lis{
		state:     StateFile(statefile),
		backlight: backlight,
		input:     make(chan struct{}),
		idle:      make(chan struct{}),
		power:     make(chan struct{}),
		signals:   make(chan struct{}),
		errors:    make(chan error),
		idleTime:  idleTime,
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

func (l *Lis) run() error {
	var err error

	// start Listening for idle
	l.idleListener()

	for {
		select {
		case <-l.input:
			fmt.Println("handle input")
			// undim screen
			l.unDim()

			// start Listening for idle
			l.idleListener()
		case <-l.idle:
			fmt.Println("idle in run")
			// get current brightness level
			err = l.getCurrent()
			if err != nil {
				return err
			}

			// dim screen
			l.dim()

			// start Listening for input to exit idle mode
			err = l.inputListener()
			if err != nil {
				return err
			}
		case power := <-l.power:
			fmt.Println("power", power)
		case signal := <-l.signals:
			fmt.Println("signal", signal)
		case err := <-l.errors:
			fmt.Println("catch any errors", err.Error())
		}
	}
}

func (l *Lis) dim() {
	fmt.Println("dim outer")
	go l.backlight.Dim(int(l.current), 0, l.errors)
}

func (l *Lis) unDim() {
	fmt.Println("undim outer")
	go l.backlight.UnDim(0, int(l.current), l.errors)
}

func (l *Lis) inputListener() error {
	devices, err := GetInputDevices()
	if err != nil {
		return err
	}

	go devices.Wait(l.input)

	return nil
}

func (l *Lis) idleListener() {
	fmt.Println("idle")
	go l.xidle()
}

// listen for X idletime
func (l *Lis) xidle() {
	for {
		time.Sleep(time.Duration(l.idleTime) * time.Millisecond)
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
