package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

func CmdIgnoreIP(args []string) {
	checkHelp(args, "Usage: beep-analytics ignore-ip <ip> [--server URL] [--token TOKEN]")
	server, token, remaining := ParseGlobalFlags(args)
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: beep-analytics ignore-ip <ip> [--server URL] [--token TOKEN]")
		os.Exit(1)
	}
	ip := remaining[0]

	client := NewClient(server, token)
	body := map[string]string{"ip": ip}
	_, err := client.Post("/api/ignore", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("IP %s added to ignore list\n", ip)
}

func CmdUnignoreIP(args []string) {
	checkHelp(args, "Usage: beep-analytics unignore-ip <ip> [--server URL] [--token TOKEN]")
	server, token, remaining := ParseGlobalFlags(args)
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: beep-analytics unignore-ip <ip> [--server URL] [--token TOKEN]")
		os.Exit(1)
	}
	ip := remaining[0]

	client := NewClient(server, token)
	_, err := client.Delete("/api/ignore/" + ip)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("IP %s removed from ignore list\n", ip)
}

func CmdListIgnored(args []string) {
	checkHelp(args, "Usage: beep-analytics list-ignored [--server URL] [--token TOKEN]")
	server, token, _ := ParseGlobalFlags(args)

	client := NewClient(server, token)
	data, err := client.Get("/api/ignore")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var ips []string
	json.Unmarshal(data, &ips)

	if len(ips) == 0 {
		fmt.Println("No IPs ignored")
		return
	}

	for _, ip := range ips {
		fmt.Println(ip)
	}
}
