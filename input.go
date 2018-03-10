package lis

import (
	"fmt"
	"io/ioutil"

	"github.com/mikkeloscar/evdev"
)

const (
	deviceDir = "/dev/input/"
)

// this way we can compare evt.Type to defined eventtypes
const (
	evKeys = uint16(evdev.EvKeys)
	evRel  = uint16(evdev.EvRelative)
)

type inputDev struct {
	devPath string
	stop    chan struct{}
	errors  chan error
}

// InputDevs defines a map of valid input devices.
type InputDevs struct {
	devs     map[string]*inputDev
	Activity chan struct{}
	errors   chan error
}

func handleDevice(inputDevice *inputDev, activity chan struct{}) {
	dev, err := evdev.Open(inputDevice.devPath)
	if err != nil {
		inputDevice.errors <- err
		return
	}
	defer dev.Close()

	for {
		select {
		case evt := <-dev.Inbox:
			if evt.Type != evKeys && evt.Type != evRel {
				continue // not the event we are looking for
			}
			// the user is still alive
			activity <- struct{}{}
		case <-inputDevice.stop:
			return
		}
	}
}

// GetInputDevices return a InputDevs containing valid input devices.
func GetInputDevices(errors chan error) (*InputDevs, error) {
	devices := &InputDevs{
		make(map[string]*inputDev),
		make(chan struct{}),
		errors,
	}

	devNames, err := ioutil.ReadDir(deviceDir)
	if err != nil {
		return nil, err
	}

	// loop through all event devices and check if they are keyboard/mouse like
	for _, d := range devNames {
		if len(d.Name()) >= 5 && d.Name()[:5] == "event" {
			devicePath := deviceDir + d.Name()
			dev, err := evdev.Open(devicePath)
			if err != nil {
				return nil, err
			}

			name, isInput := checkDevice(dev)
			if isInput {
				// don't add the same device twice
				_, ok := devices.devs[name]
				if ok {
					dev.Close()
					continue
				}

				devices.devs[name] = &inputDev{
					devicePath,
					make(chan struct{}, 1),
					errors,
				}
			}
			// close the device, since we are not gonna use it.
			dev.Close()
		}
	}
	return devices, nil
}

// Wait monitor and wait for input events and shut down on event.
func (devices *InputDevs) Wait(heartbeat chan struct{}) {
	if len(devices.devs) == 0 {
		devices.errors <- fmt.Errorf("no devices available")
		return
	}

	for _, device := range devices.devs {
		go handleDevice(device, devices.Activity)
	}

	<-devices.Activity // wait for some activity
	// fmt.Printf("Got activity!\n")

	// stop all input device listeners
	for _, device := range devices.devs {
		device.stop <- struct{}{}
	}

	heartbeat <- struct{}{} // send heartbeat to listener
}

func checkDevice(dev *evdev.Device) (string, bool) {
	if !correctDevice(dev) {
		return "", false
	}

	return dev.Name(), true
}

// check if device is a keyboard, mouse or touchpad.
func correctDevice(dev *evdev.Device) bool {
	// check if device is a keyboard
	if dev.Test(dev.EventTypes(), evdev.EvSync, evdev.EvKeys, evdev.EvMisc, evdev.EvLed, evdev.EvRepeat) {
		return true
	}

	// check if device is a mouse
	if dev.Test(dev.EventTypes(), evdev.EvSync, evdev.EvKeys, evdev.EvRelative, evdev.EvMisc) {
		return true
	}

	// check if device is a touchpad
	if dev.Test(dev.EventTypes(), evdev.EvSync, evdev.EvKeys, evdev.EvAbsolute) {
		return true
	}

	return false
}
