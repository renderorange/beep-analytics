// internal/cli/flags.go
package cli

import (
	"flag"
	"os"
)

func ParseGlobalFlags(args []string) (server string, token string, remaining []string) {
	fs := flag.NewFlagSet("beep", flag.ContinueOnError)
	fs.String("server", "", "API server URL (or set BEEP_SERVER)")
	fs.String("token", "", "API token (or set BEEP_TOKEN)")

	fs.Usage = func() {}

	fs.Parse(args)

	server = fs.Lookup("server").Value.String()
	if server == "" {
		server = os.Getenv("BEEP_SERVER")
	}
	if server == "" {
		server = "http://localhost:8080"
	}
	token = fs.Lookup("token").Value.String()
	if token == "" {
		token = LoadToken()
	}
	remaining = fs.Args()
	return
}
