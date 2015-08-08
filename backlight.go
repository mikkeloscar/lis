package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

const (
	sysPath          = "/sys/class/backlight"
	maxBrightness    = "max_brightness"
	actualBrightness = "actual_brightness"
	brightness       = "brightness"
	dimIncrement     = 10
)

// Backlight defines a backlight class from /sys/class/backlight
type Backlight struct {
	syspath string
	Max     int
}

// NewBacklight sets up a backlight struct
func NewBacklight(syspath string) (*Backlight, error) {
	fpath := path.Join(sysPath, syspath)

	backlight := &Backlight{fpath, 0}

	_, err := backlight.ReadMax()
	if err != nil {
		return nil, err
	}

	return backlight, nil
}

// reads the value of a 'brightness' file
func readInt(fpath string) (int, error) {
	buf, err := ioutil.ReadFile(fpath)
	if err != nil {
		return 0, err
	}

	num := int(0)

	for _, v := range buf {
		if '0' <= v && v <= '9' {
			num = 10*num + int(v-'0')
		} else {
			break
		}
	}
	return num, nil
}

// ReadMax get max brightness value
func (b *Backlight) ReadMax() (int, error) {
	fpath := path.Join(b.syspath, maxBrightness)

	max, err := readInt(fpath)
	if err != nil {
		return 0, err
	}
	b.Max = max // set new max

	return max, nil
}

// Get actual brightness value
func (b *Backlight) Get() (int, error) {
	fpath := path.Join(b.syspath, actualBrightness)

	current, err := readInt(fpath)
	if err != nil {
		return 0, err
	}

	return current, nil
}

// Set brightness value
func (b *Backlight) Set(value int) error {
	if value < 0 || value > b.Max {
		return fmt.Errorf("invalid brightness value '%d'", value)
	}
	fpath := path.Join(b.syspath, brightness)
	val := strconv.Itoa(value)

	fd, err := os.Open(fpath)
	if err != nil {
		return err
	}

	fd.WriteString(val)

	err = fd.Close()
	if err != nil {
		return err
	}

	return nil
}

func (b *Backlight) Dim(start, end int) error {
	var err error
	interval := (start - end) / dimIncrement
	current := start

	for i := 0; i < dimIncrement; i++ {
		current -= interval
		err = b.Set(current)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Backlight) UnDim(start, end int) error {
	var err error
	interval := (end - start) / dimIncrement
	current := start

	for i := 0; i < dimIncrement; i++ {
		current += interval
		err = b.Set(current)
		if err != nil {
			return err
		}
	}

	return nil
}

// ActualPath gets the sys-path to actual_brightness
func (b *Backlight) ActualPath() string {
	return path.Join(b.syspath, actualBrightness)
}

// func (b *Backlight) Open() (*os.File, error) {
// 	fpath := path.Join(b.syspath, actualBrightness)
// 	file, err := os.Open(fpath)
// 	return file, err
// }

// Monitor backlight for events/changes
// func (b *Backlight) Monitor() {

// }
