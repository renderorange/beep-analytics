package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

type Site struct {
	ID     int64  `json:"id"`
	Domain string `json:"domain"`
}

func CmdAddSite(args []string) {
	checkHelp(args, "Usage: beep add-site <domain> [--server URL] [--token TOKEN]")
	server, token, remaining := ParseGlobalFlags(args)
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: beep add-site <domain> [--server URL] [--token TOKEN]")
		os.Exit(1)
	}
	domain := remaining[0]

	client := NewClient(server, token)
	body := map[string]string{"domain": domain}
	_, err := client.Post("/api/sites", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Site %s added\n", domain)
}

func CmdRemoveSite(args []string) {
	checkHelp(args, "Usage: beep remove-site <domain> [--server URL] [--token TOKEN]")
	server, token, remaining := ParseGlobalFlags(args)
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: beep remove-site <domain> [--server URL] [--token TOKEN]")
		os.Exit(1)
	}
	domain := remaining[0]

	client := NewClient(server, token)
	_, err := client.Delete("/api/sites/" + domain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Site %s removed\n", domain)
}

func CmdListSites(args []string) {
	checkHelp(args, "Usage: beep list-sites [--server URL] [--token TOKEN]")
	server, token, _ := ParseGlobalFlags(args)

	client := NewClient(server, token)
	data, err := client.Get("/api/sites")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var sites []Site
	json.Unmarshal(data, &sites)

	if len(sites) == 0 {
		fmt.Println("No sites registered")
		return
	}

	for _, s := range sites {
		fmt.Printf("%s\n", s.Domain)
	}
}
