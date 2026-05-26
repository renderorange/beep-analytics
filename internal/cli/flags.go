package cli

import (
	"fmt"
	"os"
	"strings"
)

func checkHelp(args []string, usage string) {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(0)
		}
	}
}

func ParseGlobalFlags(args []string) (server string, token string, remaining []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--server=") {
			server = arg[len("--server="):]
		} else if strings.HasPrefix(arg, "--token=") {
			token = arg[len("--token="):]
		} else if arg == "--server" && i+1 < len(args) {
			server = args[i+1]
			i++
		} else if arg == "--token" && i+1 < len(args) {
			token = args[i+1]
			i++
		} else {
			remaining = append(remaining, arg)
		}
	}

	if server == "" {
		server = os.Getenv("BEEP_SERVER")
	}
	if server == "" {
		server = "http://localhost:8080"
	}
	if token == "" {
		token = LoadToken()
	}
	return
}
