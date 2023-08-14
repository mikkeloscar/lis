package lis

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// Lis defines the core state of the lis daemon.
type Lis struct {
	current   uint16        // current brightness value
	idleMode  bool          // true if in idle mode
	state     StateFile     // state file
	backlight *Backlight    // backlight
	input     chan struct{} // input channel used to notify about activity when in idle mode
	idle      chan struct{} // idle channel used when user is idle
	power     chan struct{} // power channel used to notify about power changes (AC/Battery)
	// stop      <-chan struct{} // stop channel used to stop the lis main loop
	errors   chan error  // errors channel
	IPC      chan IPCCmd // ipc channel used to communicate with the IPC server
	idleTime uint        // idle time in minutes
}

// NewLis creates a new Lis instance.
func NewLis(config *Config) (*Lis, error) {
	var backlightName string
	switch config.Backlight {
	case "intel":
		backlightName = "intel_backlight"
	case "amdgpu":
		backlightName = "amdgpu_bl1"
	default:
		return nil, fmt.Errorf("backlight: %s not supported", config.Backlight)
	}

	backlight, err := NewBacklight(backlightName)
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
		errors:    make(chan error),
		IPC:       make(chan IPCCmd),
		idleTime:  config.IdleTime,
	}, nil
}

// load state from stateFile.
func (l *Lis) loadState() error {
	v, err := l.state.Read()
	if err != nil {
		// if state files was not found set brightness to max value
		if os.IsNotExist(err) {
			max, err := l.backlight.ReadMax()
			if err != nil {
				return err
			}
			v = uint16(max)
		} else {
			return err
		}
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

// Run runs the lis main loop.
func (l *Lis) Run(ctx context.Context) error {
	// load initial state
	err := l.loadState()
	if err != nil {
		return err
	}

	dbus, err := NewDBusHandler(l)
	if err != nil {
		return err
	}

	go dbus.Run(l.errors)
	defer dbus.Close()

	// start IPC server
	ipc, err := NewIPCServer()
	if err != nil {
		return err
	}

	go ipc.Run(l.IPC, l.errors)
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
			log.Info("Entering idle mode")

			// dim screen
			l.dim()
			l.idleMode = true

			// start Listening for input to exit idle mode
			err = l.inputListener()
			if err != nil {
				log.Error(err)
				continue
			}
		case power := <-l.power:
			fmt.Println("power", power)
		case ipc := <-l.IPC:
			switch ipc.typ {
			case IPCSet:
				err = l.SetPercent(ipc.val.(float64))
				if err != nil {
					log.Errorf("Failed to set brightness value: %v", err)
				}
				ipc.resp <- err
			case IPCSetUp, IPCSetDown:
				current, err := l.GetPercent()
				if err != nil {
					log.Errorf("Failed to get brightness value: %v", err)
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
						log.Errorf("Failed to set brightness value: %v", err)
					}
				}

				ipc.resp <- err
			case IPCStatus:
				val, err := l.GetPercent()
				if err != nil {
					log.Errorf("Failed to get brightness value: %s", err)
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
		case <-ctx.Done():
			return l.storeState()
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
	l.current = uint16(val)
	return l.backlight.Set(val)
}

// dim screen.
func (l *Lis) dim() {
	log.Infof("Dimming screen from brightness level %d to %d", l.current, 0)
	go l.backlight.Dim(int(l.current), 0, l.errors)
}

// undim screen.
func (l *Lis) unDim() {
	log.Infof("Undimming screen to brightness level %d to %d", 0, l.current)
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

		log.Infof("Idling for %s", time.Duration(idleTime)*time.Millisecond)

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
