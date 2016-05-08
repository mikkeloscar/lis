package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Lis defines the core state of the lis daemon.
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
	IPC       chan IPCCmd    // ipc channel used to communicate with the IPC server
	idleTime  uint           // idle time in minutes
}

// NewLis creates a new Lis instance.
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
		IPC:       make(chan IPCCmd),
		idleTime:  config.IdleTime,
	}, nil
}

// load state from stateFile.
func (l *Lis) loadState() error {
	v, err := l.state.Read()
	if err != nil {
		return err
	}

	l.current = v

	err = l.backlight.Set(int(l.current))
	if err != nil {
		return err
	}

	return nil
}

// store current state in stateFile.
func (l *Lis) storeState() error {
	if l.idleMode {
		err := l.state.Write(l.current)
		if err != nil {
			return err
		}

		return nil
	}

	err := l.getCurrent()
	if err != nil {
		return err
	}

	err = l.state.Write(l.current)
	if err != nil {
		return err
	}

	return nil
}

// get current brightness level.
func (l *Lis) getCurrent() error {
	v, err := l.backlight.Get()
	if err != nil {
		return err
	}

	l.current = uint16(v)

	return nil
}

// lis main loop.
func (l *Lis) run() {
	var err error

	dbus, err := NewDBusHandler(l)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	go dbus.Run(l.errors)

	// start IPC server
	ipc, err := NewIPCServer()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	go ipc.Run(l)
	defer ipc.Close()

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
			case syscall.SIGINT:
				exit := 0

				err = l.storeState()
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					exit = 1
				}

				// shutdown ipc server
				err = ipc.Close()
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					exit = 1
				}

				os.Exit(exit)
			}
		case ipc := <-l.IPC:
			switch ipc.typ {
			case IPCSet:
				err = l.SetPercent(ipc.val.(float64))
				if err != nil {
					err = fmt.Errorf("failed to set brightness value: %s", err)
					log.Errorf(err.Error())
				}
				ipc.resp <- err
			case IPCSetUp, IPCSetDown:
				current, err := l.GetPercent()
				if err != nil {
					err = fmt.Errorf("failed to get brightness value: %s", err)
					log.Errorf(err.Error())
				} else {
					var value float64
					switch ipc.typ {
					case IPCSetUp:
						value = clampPct(current + ipc.val.(float64))
					case IPCSetDown:
						value = clampPct(current - ipc.val.(float64))
					}
					err = l.SetPercent(value)
					if err != nil {
						err = fmt.Errorf("failed to set brightness value: %s", err)
						log.Errorf(err.Error())
					}
				}

				ipc.resp <- err
			case IPCStatus:
				val, err := l.GetPercent()
				if err != nil {
					err = fmt.Errorf("failed to get brightness value: %s", err)
					log.Errorf(err.Error())
					ipc.resp <- err
				} else {
					ipc.resp <- val
				}
			case IPCDPMSOn:
			case IPCDPMSOff:
			}

		case err := <-l.errors:
			// Write error to stderr
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
}

// GetPercent gets the current backlight value as a percent value.
// (max/current).
func (l *Lis) GetPercent() (float64, error) {
	val, err := l.backlight.Get()
	if err != nil {
		return 0, err
	}

	return float64(val) / float64(l.backlight.Max), nil
}

// SetPercent sets the current value from a percent value. (max * value).
func (l *Lis) SetPercent(value float64) error {
	if value > 1 || value < 0 {
		return fmt.Errorf("invalid percent value: %f", value)
	}

	val := int(float64(l.backlight.Max) * value)
	return l.backlight.Set(val)
}

// dim screen.
func (l *Lis) dim() {
	fmt.Printf("dimming screen from brightness level: %d\n", l.current)
	go l.backlight.Dim(int(l.current), 0, l.errors)
}

// undim screen.
func (l *Lis) unDim() {
	fmt.Printf("undimming screen to brightness level: %d\n", l.current)
	go l.backlight.UnDim(0, int(l.current), l.errors)
}

// listen for input activity.
func (l *Lis) inputListener() error {
	devices, err := GetInputDevices(l.errors)
	if err != nil {
		return err
	}

	go devices.Wait(l.input)

	return nil
}

// listen for user idling.
func (l *Lis) idleListener() {
	go l.xidle()
}

// listen for X idletime.
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

func clampPct(value float64) float64 {
	if value > 1 {
		return 1
	}

	if value < 0 {
		return 0
	}

	return value
}
