package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		client := &IPCClient{}
		var err error
		switch os.Args[1] {
		case "set":
			if len(os.Args) < 3 {
				// error
			}
			err = client.Set(os.Args[2])
		case "status":
			var resp string
			resp, err = client.Status()
			if err == nil {
				fmt.Println(resp)
			}
		case "dpms":
			if len(os.Args) < 3 {
				// error
			}
			err = client.DPMS(os.Args[2])
		default:
			// invalid
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		return
	}
}
