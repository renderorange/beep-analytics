package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/adventurehound/beep/internal/api"
	"github.com/adventurehound/beep/internal/cli"
	"github.com/adventurehound/beep/internal/db"
	"github.com/adventurehound/beep/internal/geoip"
)

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.String("port", "8080", "Port to listen on")
	dbPath := fs.String("db", "beep.db", "Path to SQLite database")
	geoipPath := fs.String("geoip", "", "Path to GeoLite2 CSV directory (optional)")
	fs.Parse(args)

	database, err := db.Open(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	geo, err := geoip.New(*geoipPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading GeoIP: %v\n", err)
		os.Exit(1)
	}

	addr := ":" + *port
	server := api.NewServer(database, geo, addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func cmdAddSite(args []string)       { cli.CmdAddSite(args) }
func cmdRemoveSite(args []string)    { cli.CmdRemoveSite(args) }
func cmdListSites(args []string)     { cli.CmdListSites(args) }
func cmdIgnoreIP(args []string)      { cli.CmdIgnoreIP(args) }
func cmdUnignoreIP(args []string)    { cli.CmdUnignoreIP(args) }
func cmdListIgnored(args []string)   { cli.CmdListIgnored(args) }
func cmdGenerateToken(args []string) { cli.CmdGenerateToken(args) }
func cmdRevokeToken(args []string)   { cli.CmdRevokeToken(args) }
func cmdStats(args []string)         { cli.CmdStats(args) }
