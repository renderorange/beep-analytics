package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func CmdGenerateToken(args []string) {
	checkHelp(args, "Usage: beep-analytics generate-token [--server URL] [--token TOKEN]")
	server, token, _ := ParseGlobalFlags(args)

	client := NewClient(server, token)
	data, err := client.Post("/api/tokens/generate", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var result struct {
		Token string `json:"token"`
		ID    int64  `json:"id"`
	}
	json.Unmarshal(data, &result)

	fmt.Printf("Token ID: %d\n", result.ID)
	fmt.Printf("Token: %s\n", result.Token)
	fmt.Println("\nSave this token securely. It cannot be retrieved again.")
}

func CmdRevokeToken(args []string) {
	checkHelp(args, "Usage: beep-analytics revoke-token <id> [--server URL] [--token TOKEN]")
	server, token, remaining := ParseGlobalFlags(args)
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: beep-analytics revoke-token <id> [--server URL] [--token TOKEN]")
		os.Exit(1)
	}
	id := remaining[0]

	_, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid token ID")
		os.Exit(1)
	}

	client := NewClient(server, token)
	_, err = client.Delete("/api/tokens/" + id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Token %s revoked\n", id)
}
