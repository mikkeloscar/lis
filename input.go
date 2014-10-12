package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jteeuwen/evdev"
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
}

// InputDevs defines a map of valid input devices
type InputDevs struct {
	devs     map[string]*inputDev
	Activity chan struct{}
}

func handleDevice(inputDevice *inputDev, activity chan struct{}) {
	dev, err := evdev.Open(inputDevice.devPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
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

// GetInputDevices return a InputDevs containing valid input devices
func GetInputDevices() (*InputDevs, error) {
	devices := &InputDevs{make(map[string]*inputDev), make(chan struct{})}

	devNames, err := ioutil.ReadDir(deviceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return nil, err
	}

	// loop through all event devices and check if they are keyboard/mouse like
	for _, d := range devNames {
		if len(d.Name()) >= 5 && d.Name()[:5] == "event" {
			devicePath := deviceDir + d.Name()
			dev, err := evdev.Open(devicePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}

			name, isInput := checkDevice(dev)
			if isInput {
				// TODO needed? is the same device represented more than
				// one time?
				_, ok := devices.devs[name]
				if ok {
					dev.Close()
					continue
				}

				devices.devs[name] = &inputDev{
					devicePath,
					make(chan struct{}),
				}
			}
			// close the device, since we are not gonna use it.
			dev.Close()
		}
	}
	return devices, nil
}

// Listen for input events and shut down on event.
func (devices *InputDevs) Listen(heartbeat chan struct{}) {
	if len(devices.devs) == 0 {
		fmt.Fprintf(os.Stderr, "no devices available\n")
		return
	}

	for _, device := range devices.devs {
		go handleDevice(device, devices.Activity)
	}

	<-devices.Activity // wait for some activity
	fmt.Printf("Got activity!\n")

	// stop all input device listeners
	for _, device := range devices.devs {
		device.stop <- struct{}{}
	}

	heartbeat <- struct{}{} // send heartbeat to main
}

func checkDevice(dev *evdev.Device) (string, bool) {
	if !correctDevice(dev) {
		return "", false
	}

	return dev.Name(), true
}

// TODO refactor maybe
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
	if dev.Test(dev.EventTypes(), evdev.EvSync, evdev.EvKeys, evdev.EvRelative, evdev.EvAbsolute) {
		return true
	}

	return false
}

func main() {
	devices, err := GetInputDevices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	heartbeat := make(chan struct{})

	go devices.Listen(heartbeat)

	select {
	case <-heartbeat:
		fmt.Println("hello")
		return
	}
}
