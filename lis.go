package main

import (
	"fmt"
	"os"
	"time"
)

type lis struct {
	current   uint16        // current brightness value
	state     StateFile     // state file
	backlight *Backlight    // backlight
	input     chan struct{} // input channel used to notify about activity when in idle mode
	idle      chan struct{} // idle channel used when user is idle
	power     chan struct{} // power channel used to notify about power changes (AC/Battery)
	signals   chan struct{} // signals channel used to handle external signals
	idleTime  uint          // idle time in minutes
}

// load state from stateFile
func (l *lis) loadState() error {
	v, err := l.state.Read()
	if err != nil {
		return err
	}

	l.current = v

	return nil
}

// store current state in stateFile
func (l *lis) storeState() error {
	return l.state.Write(l.current)
}

// get current brightness level
func (l *lis) getCurrent() error {
	v, err := l.backlight.Get()
	if err != nil {
		return err
	}

	l.current = uint16(v)

	return nil
}

func (l *lis) run() error {
	var err error

	// start listening for idle
	l.idleListener()

	for {
		select {
		case <-l.input:
			fmt.Println("handle input")
			// undim screen
			err = l.unDim()
			if err != nil {
				return err
			}

			// start listening for idle
			l.idleListener()
		case <-l.idle:
			fmt.Println("idle")
			// get current brightness level
			err = l.getCurrent()
			if err != nil {
				return err
			}

			// dim screen
			err = l.dim()
			if err != nil {
				return err
			}

			// start listening for input to exit idle mode
			err = l.inputListener()
			if err != nil {
				return err
			}
		case power := <-l.power:
			fmt.Println("power", power)
		case signal := <-l.signals:
			fmt.Println("signal", signal)
		}
	}
}

func (l *lis) dim() error {
	return l.backlight.Dim(int(l.current), 0)
}

func (l *lis) unDim() error {
	return l.backlight.UnDim(0, int(l.current))
}

func (l *lis) inputListener() error {
	devices, err := GetInputDevices()
	if err != nil {
		return err
	}

	go devices.Wait(l.input)

	return nil
}

func (l *lis) idleListener() {
	go l.xidle()
}

// TODO better error handling
func (l *lis) xidle() {
	for {
		time.Sleep(time.Duration(l.idleTime) * time.Minute)
		idleTime, err := XIdle()
		if err != nil {
			// TODO better
			fmt.Println("error in XIdle: ", err.Error())
			continue
		}

		if idleTime >= l.idleTime*60*60*1000 {
			l.idle <- struct{}{}
			break
		}
	}
}

func main() {
	devices, err := GetInputDevices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	heartbeat := make(chan struct{})

	go devices.Wait(heartbeat)

	for {

		// select {
		// case <-heartbeat:
		<-heartbeat
		fmt.Println("hello")
		// return
		// }
	}
}

// // monitor, _ := NewMonitor()
// b, _ := NewBacklight("intel_backlight")

// watcher, err := inotify.NewWatcher()
// if err != nil {
// 	fmt.Println("Error")
// }

// err = watcher.Watch(b.ActualPath())
// if err != nil {
// 	fmt.Println("Error2")
// }

// // f, _ := os.Open(b.ActualPath())

// // fd := int32(f.Fd())

// // defer f.Close()

// // if err != nil {
// // 	return
// // }

// // defer monitor.Close()

// // u := udev.NewUdev()
// // defer u.Unref()

// // mon := udev.NewMonitorFromNetlink(u, "udev")
// // mon.AddFilter("backlight", "")
// // mon.EnableReceiving()

// // fd := int32(mon.Fd())

// // fmt.Println(syscall.SetNonblock(int(fd), true))

// // monitor.Register(fd, "actual_brightness")

// // monitor.Poll()

// for {
// 	select {
// 	case ev := <-watcher.Event:
// 		fmt.Println("event:", ev)
// 	case err := <-watcher.Error:
// 		fmt.Println("error:", err)
// 		// case ev := <-watcher.Event:
// 		// 	fmt.Printf("Event: %s\n", ev)
// 		// 	var buf [1024]byte

// 		// 	for {
// 		// 		n, e := syscall.Read(int(ev.Fd), buf[:])
// 		// 		if n > 0 {
// 		// 			fmt.Printf("got something: %#v\n", buf[:n])
// 		// 		} else {
// 		// 			break
// 		// 		}

// 		// 		fmt.Printf("error: %s\n", e)
// 		// 	}
// 		// dev := mon.ReceiveDevice()
// 		// if dev != nil {
// 		// 	fmt.Printf("Got Device\n")
// 		// 	fmt.Printf("   Node: %#v\n", dev.DevNode())
// 		// 	fmt.Printf("   Subsystem: %#v\n", dev.Subsystem())
// 		// 	fmt.Printf("   DevType: %#v\n", dev.DevType())
// 		// 	fmt.Printf("   Action: %#v\n", dev.Action())
// 		// } else {
// 		// 	fmt.Printf("No Device from received_device(). An error occured.\n")
// 		// }
// 	}
// }
// // e := u.NewEnumerate()
// // defer e.Unref()

// // e.AddMatchSubsystem("backlight")
// // e.ScanDevices()

// // for device := e.First(); !device.IsNil(); device = device.Next() {
// // 	path := device.Name()
// // 	// fmt.Println(path)
// // 	dev := u.DeviceFromSysPath(path)

// // 	fmt.Printf("Device Node Path: %s\n", dev.DevNode())
// // }

// }
