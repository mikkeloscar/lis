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
	inhibit *os.File
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

	d.Signal = d.login1.Subscribe("PrepareForSleep")

	// take inhibit lock
	d.takeLock(errCh)

	for {
		signal := <-d.Signal
		switch signal.Name {
		case "org.freedesktop.login1.Manager.PrepareForSleep":
			prepareForSleep := signal.Body[0].(bool)
			if prepareForSleep { // prepare for suspend
				err = d.lis.storeState()
				if err != nil {
					errCh <- err
				}

				err = d.inhibit.Close()
				if err != nil {
					errCh <- err
				}

				d.inhibit = nil
			} else { // go back from suspend
				d.lis.loadState()
				// take new lock
				d.takeLock(errCh)
			}
		}

	}
}

// take inhibitor lock if not already taken
func (d *DBusHandler) takeLock(errCh chan error) {
	var err error
	if d.inhibit == nil {
		d.inhibit, err = d.login1.Inhibit("sleep", "Backlight daemon", "Saving backlight level", "delay")
		if err != nil {
			errCh <- err
		}
	}
}

// Close closes the inhibit file
func (d *DBusHandler) Close() {
	d.closeCh <- struct{}{}
}
