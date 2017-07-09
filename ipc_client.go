package lis

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

const socket = "/var/run/lis.sock"

var setPatt = regexp.MustCompile(`(\+|-)?(\d+)%`)

// IPCClient defines an IPC client for communicating with the lis IPC server.
type IPCClient struct {
	net.Conn
}

// RPC sends a message to the IPC server and handles the response.
func (i *IPCClient) RPC(msg string, args ...interface{}) (interface{}, error) {
	var err error
	i.Conn, err = net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}
	defer i.Close()

	_, err = i.Write([]byte(fmt.Sprintf(msg+"\n", args)))
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(i)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	resp := strings.Split(line[:len(line)-1], " ")
	if len(resp) > 0 {
		if resp[0] == "OK" {
			if len(resp) > 1 {
				return resp[1], nil
			}
			return nil, nil
		}

		if resp[0] == "ERROR" {
			if len(resp) > 1 {
				return nil, fmt.Errorf(resp[1])
			}
		}
	}

	return nil, fmt.Errorf("invalid response: %s", line[:len(line)-1])
}

// Set sets the brightness value via IPC.
func (i *IPCClient) Set(value string) error {
	match := setPatt.FindStringSubmatch(value)
	if len(match) == 0 {
		return fmt.Errorf("invalid SET argument: %s", value)
	}

	intVal, err := strconv.ParseInt(match[2], 10, 8)
	if err != nil {
		return err
	}

	if intVal < 0 || intVal > 100 {
		return fmt.Errorf("invalid SET argument: %s", value)
	}

	_, err = i.RPC("SET %s", value)
	return err
}

// Status gets the brightness status via IPC.
func (i *IPCClient) Status() (string, error) {
	val, err := i.RPC("STATUS")
	if err != nil {
		return "", err
	}
	return val.(string), err
}

// DPMS sets the enables/disables DPMS via IPC.
func (i *IPCClient) DPMS(value string) error {
	// TODO: check value
	switch value {
	case "on", "off":
		_, err := i.RPC("DPMS %s", value)
		return err
	default:
		return fmt.Errorf("invalid value '%s', must be one of 'on, off'", value)
	}
}
