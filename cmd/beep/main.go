package main

import (
    "fmt"
    "os"
)

func usage() {
    fmt.Fprintf(os.Stderr, `Usage: beep <command> [options]

Commands:
  serve                Start the tracking server
  add-site <domain>    Register a site to track
  remove-site <domain> Remove a site
  list-sites           List tracked sites
  ignore-ip <ip>       Add IP to ignore list
  unignore-ip <ip>     Remove IP from ignore list
  list-ignored         List ignored IPs
  generate-token       Create a new API token
  revoke-token <id>    Revoke a token
  stats                Show tracking stats
  version              Show version

Run 'beep <command> --help' for command-specific help.
`)
    os.Exit(1)
}

func main() {
    if len(os.Args) < 2 {
        usage()
    }

    switch os.Args[1] {
    case "serve":
        cmdServe(os.Args[2:])
    case "add-site":
        cmdAddSite(os.Args[2:])
    case "remove-site":
        cmdRemoveSite(os.Args[2:])
    case "list-sites":
        cmdListSites(os.Args[2:])
    case "ignore-ip":
        cmdIgnoreIP(os.Args[2:])
    case "unignore-ip":
        cmdUnignoreIP(os.Args[2:])
    case "list-ignored":
        cmdListIgnored(os.Args[2:])
    case "generate-token":
        cmdGenerateToken(os.Args[2:])
    case "revoke-token":
        cmdRevokeToken(os.Args[2:])
    case "stats":
        cmdStats(os.Args[2:])
    case "version":
        fmt.Println("beep v0.1.0")
    case "help", "--help", "-h":
        usage()
    default:
        fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
        usage()
    }
}
