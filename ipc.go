package main

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

var setPatt = regexp.MustCompile(`(\+|-)?(\d+)%`)

const socket = "/var/run/lis.sock"

type IPCCmdType int

const (
	IPCSet IPCCmdType = iota
	IPCSetUp
	IPCSetDown
	IPCStatus
	IPCDPMSOn
	IPCDPMSOff
)

type IPCCmd struct {
	typ  IPCCmdType
	val  interface{}
	resp chan interface{}
}

type client struct {
	net.Conn
	lis *Lis
}

// Ok sends an OK response to the client.
func (c *client) Ok() {
	fmt.Fprintf(c, "OK\n")
}

// OkMsg sends an OK response wth a message to the client.
func (c *client) OkMsg(msg string, args ...interface{}) {
	fmt.Fprintf(c, "OK "+msg+"\n", args...)
}

// Errorf sends an error response to the client.
func (c *client) Errorf(msg string, args ...interface{}) {
	fmt.Fprintf(c, "ERROR "+msg+"\n", args...)
}

func IPCServer(lis *Lis) {
	ln, err := net.Listen("unix", socket)
	if err != nil {
		lis.errors <- fmt.Errorf("failed to start IPC socker at %s", socket)
		return
	}

	log.Infof("IPC server listening on socket: %s", socket)
	for {
		conn, err := ln.Accept()
		if err != nil {
			lis.errors <- fmt.Errorf("accept error: %s", err)
		}
		go handleConnection(&client{conn, lis})
	}
}

func handleConnection(client *client) {
	defer client.Close()

	reader := bufio.NewReader(client)
	line, err := reader.ReadString('\n')
	if err != nil {
		client.lis.errors <- fmt.Errorf("unable to read from client: %s", err)
		return
	}

	cmd, args := parseCmd(line[:len(line)-1])
	ipcCmd := IPCCmd{resp: make(chan interface{})}
	switch cmd {
	case "SET":
		if len(args) == 0 {
			client.Errorf("Missing SET argument")
			break
		}

		match := setPatt.FindStringSubmatch(args[0])
		if len(match) == 0 {
			client.Errorf("Invalid")
			break
		}

		value, err := strconv.ParseInt(match[2], 10, 8)
		if err != nil {
			client.Errorf("Failed to parse value: %s", args[0])
			break
		}

		ipcCmd.val = float64(value) / 100

		switch match[1] {
		case "":
			ipcCmd.typ = IPCSet
		case "+":
			ipcCmd.typ = IPCSetUp
		case "-":
			ipcCmd.typ = IPCSetDown
		}

		client.lis.IPC <- ipcCmd
		ok := <-ipcCmd.resp
		if ok != nil {
			err = ok.(error)
			client.Errorf(err.Error())
		} else {
			client.Ok()
		}
	case "STATUS":
		ipcCmd.typ = IPCStatus
		client.lis.IPC <- ipcCmd

		status := <-ipcCmd.resp

		switch v := status.(type) {
		case error:
			client.Errorf(v.Error())
		case float64:
			client.OkMsg("%d%%", int(v*100))
		}
	case "DPMS":
		if len(args) == 0 {
			client.Errorf("Missing DPMS argument")
			break
		}

		switch args[0] {
		case "ON":
			// enable DPMS
			ipcCmd.typ = IPCDPMSOn
			client.lis.IPC <- ipcCmd
			ok := <-ipcCmd.resp
			if ok.(bool) {
				client.Ok()
			} else {
				client.Errorf("Failed to set DPMS ON")
			}
		case "OFF":
			// disable DPMS
			ipcCmd.typ = IPCDPMSOff
			client.lis.IPC <- ipcCmd
			ok := <-ipcCmd.resp
			if ok.(bool) {
				client.Ok()
			} else {
				client.Errorf("Failed to set DPMS OFF")
			}
		default:
			client.Errorf("Invalid DPMS argument: %s", args[0])
		}
	default:
		client.Errorf("Invalid command: %s", cmd)
	}

	close(ipcCmd.resp)
}

func parseCmd(line string) (string, []string) {
	cmd := strings.Split(line, " ")
	if len(cmd) > 1 {
		return cmd[0], cmd[1:]
	}

	return cmd[0], nil
}
