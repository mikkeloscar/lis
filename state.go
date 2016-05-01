package main

import (
	"encoding/binary"
	"os"
)

// StateFile is a path to the lis state file.
type StateFile string

// Read value from stateFile.
func (s StateFile) Read() (uint16, error) {
	file, err := os.Open(string(s))
	if err != nil {
		return 0, err
	}
	defer file.Close()

	data := make([]byte, 2)

	_, err = file.Read(data)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint16(data), nil
}

// Write value to stateFile.
func (s StateFile) Write(value uint16) error {
	file, err := os.OpenFile(string(s), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, value)

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
