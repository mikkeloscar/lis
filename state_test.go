package main

import (
	"os"
	"testing"
)

func TestWrite(t *testing.T) {
	state := StateFile("test")

	err := state.Write(100)
	if err != nil {
		t.Errorf("failed to write state: %s", err.Error())
	}
}

func TestRead(t *testing.T) {
	state := StateFile("test")

	v, err := state.Read()
	if err != nil {
		t.Errorf("should not cause error")
	}

	if v != 100 {
		t.Errorf("should be %d, was %d", 100, v)
	}

	err = os.Remove("test")
	if err != nil {
		t.Errorf("error when removing test state file")
	}
}
