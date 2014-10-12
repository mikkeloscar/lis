package main

import (
	"errors"
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
)

// Backlight defines a state for the screen backlight
type Backlight struct {
	interf  string
	Max     int
	Current int
	Default int // brightness value set by the user
}

// InitBacklight sets up a backlight state struct
func InitBacklight(interf string, def int) (*Backlight, error) {
	fpath := path.Join(sysPath, interf)

	backlight := &Backlight{fpath, 0, 0, 0}

	backlight.GetMax()
	backlight.SetBacklight(def)

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

// GetMax get max brightness value
func (b *Backlight) GetMax() (int, error) {
	fpath := path.Join(b.interf, maxBrightness)

	max, err := readInt(fpath)
	if err != nil {
		return 0, err
	}
	b.Max = max // set new max

	return max, nil
}

// GetActual get actual brightness value
func (b *Backlight) GetActual() (int, error) {
	fpath := path.Join(b.interf, actualBrightness)

	current, err := readInt(fpath)
	if err != nil {
		return 0, err
	}
	b.Current = current // set new current

	return current, nil
}

// SetBacklight set brightness value
func (b *Backlight) SetBacklight(value int) error {
	if value < 0 {
		return errors.New("invalid value")
	}
	fpath := path.Join(b.interf, brightness)
	val := strconv.Itoa(value)

	fd, err := os.Open(fpath)
	if err != nil {
		return err
	}

	fd.WriteString(val)

	fd.Close()

	// set current and default value
	b.Current = value
	b.Default = value

	return nil
}

// Monitor backlight for events/changes
// func (b *Backlight) Monitor() {

// }
