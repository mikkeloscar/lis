package main

import (
	"os"

	"github.com/godbus/dbus"
	"github.com/mikkeloscar/go-systemd/login1"
)

// DBusHandler handles a DBus connection to receive signal on suspend
type DBusHandler struct {
	login1  *login1.Conn
	Signal  chan *dbus.Signal
	closeCh chan struct{}
	lis     *Lis
}

// NewDBusHandler initializes a new DBusHandler
func NewDBusHandler(lis *Lis) (*DBusHandler, error) {
	l, err := login1.New()
	if err != nil {
		return nil, err
	}

	return &DBusHandler{
		login1:  l,
		closeCh: make(chan struct{}, 1),
		lis:     lis,
	}, nil

}

// Run starts the DBusHandler event loop
func (d *DBusHandler) Run(errCh chan error) {
	var err error
	var file *os.File

	d.Signal = d.login1.Subscribe("PrepareForSleep")

	for {
		if file == nil {
			file, err = d.login1.Inhibit("sleep", "Backlight daemon", "Saving backlight level", "delay")
			if err != nil {
				errCh <- err
			}
		}
		signal := <-d.Signal
		switch signal.Name {
		case "org.freedesktop.login1.Manager.PrepareForSleep":
			prepareForSleep := signal.Body[0].(bool)
			if prepareForSleep { // prepare for suspend
				err = d.lis.storeState()
				if err != nil {
					errCh <- err
				}

				file.Close()
				file = nil
			} else { // go back from suspend
				d.lis.loadState()
			}
		}

	}
}

// Close closes the inhibit file
func (d *DBusHandler) Close() {
	d.closeCh <- struct{}{}
}
