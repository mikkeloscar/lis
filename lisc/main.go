package main

import (
	"fmt"
	"os"
)

func usage(exit int) {
	usage := `Usage: lisc [COMMAND] ...

Control lis daemon.

  COMMANDS:
    set <+|-value%>    set/increase/decrease brightness level
    status	   get current brightness level
    dmps <on|off>  set dpms on/off

  OPTIONS:
    -h, --help     display this help mesage
`

	fmt.Printf("%s", usage)
	os.Exit(exit)
}

func main() {
	if len(os.Args) > 1 {
		client := &IPCClient{}
		var err error
		switch os.Args[1] {
		case "set":
			if len(os.Args) < 3 {
				// invalid command
				usage(1)
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
				// invalid command
				usage(1)
			}
			err = client.DPMS(os.Args[2])
		case "-h", "--help":
			usage(0)
		default:
			// invalid command
			usage(1)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		return
	}
	// invalid command
	usage(1)
}
