// internal/cli/flags.go
package cli

import (
	"flag"
)

func ParseGlobalFlags(args []string) (server string, token string, remaining []string) {
	fs := flag.NewFlagSet("beep", flag.ContinueOnError)
	fs.String("server", "http://localhost:8080", "API server URL")
	fs.String("token", "", "API token (or set SMALLEST_TRACKER_TOKEN)")

	fs.Usage = func() {}

	fs.Parse(args)

	server = fs.Lookup("server").Value.String()
	token = fs.Lookup("token").Value.String()
	if token == "" {
		token = LoadToken()
	}
	remaining = fs.Args()
	return
}
