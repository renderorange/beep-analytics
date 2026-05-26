package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/adventurehound/beep/internal/models"
)

func CmdStats(args []string) {
	checkHelp(args, "Usage: beep stats [--site DOMAIN] [--last DURATION] [--from TIME] [--to TIME] [--verbose] [--server URL] [--token TOKEN]")
	server, token, remaining := ParseGlobalFlags(args)

	var site, last, from, to string
	var verbose bool

	for i := 0; i < len(remaining); i++ {
		switch remaining[i] {
		case "--site":
			if i+1 < len(remaining) {
				site = remaining[i+1]
				i++
			}
		case "--last":
			if i+1 < len(remaining) {
				last = remaining[i+1]
				i++
			}
		case "--from":
			if i+1 < len(remaining) {
				from = remaining[i+1]
				i++
			}
		case "--to":
			if i+1 < len(remaining) {
				to = remaining[i+1]
				i++
			}
		case "--verbose", "-v":
			verbose = true
		}
	}

	params := url.Values{}
	if site != "" {
		params.Set("site", site)
	}
	if last != "" {
		params.Set("last", last)
	}
	if from != "" {
		params.Set("from", from)
	}
	if to != "" {
		params.Set("to", to)
	}
	if verbose {
		params.Set("verbose", "true")
	}

	client := NewClient(server, token)
	data, err := client.Get("/api/stats?" + params.Encode())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		displayVerboseStats(data, site)
	} else {
		displayAggregateStats(data, site)
	}
}

func displayAggregateStats(data []byte, filterSite string) {
	var stats []models.StatsRow
	json.Unmarshal(data, &stats)

	if len(stats) == 0 {
		fmt.Println("No data")
		return
	}

	grouped := make(map[string][]models.StatsRow)
	for _, s := range stats {
		grouped[s.Site] = append(grouped[s.Site], s)
	}

	showSite := filterSite == ""

	for site, rows := range grouped {
		if showSite {
			fmt.Printf("\n=== %s ===\n", site)
		}
		fmt.Printf("%-20s %-20s %s\n", "IP", "Path", "Count")
		for _, r := range rows {
			fmt.Printf("%-20s %-20s %d\n", r.IP, r.Path, r.Count)
		}
	}
}

func displayVerboseStats(data []byte, filterSite string) {
	var stats []models.StatsRow
	json.Unmarshal(data, &stats)

	if len(stats) == 0 {
		fmt.Println("No data")
		return
	}

	grouped := make(map[string][]models.StatsRow)
	for _, s := range stats {
		grouped[s.Site] = append(grouped[s.Site], s)
	}

	showSite := filterSite == ""

	for site, rows := range grouped {
		if showSite {
			fmt.Printf("\n=== %s ===\n", site)
		}
		fmt.Printf("%-20s %-8s %-15s %-15s %-10s %-10s %-15s %-20s %s\n",
			"IP", "Country", "Region", "City", "Browser", "OS", "Path", "Referrer", "Time")
		for _, r := range rows {
			time := r.Time
			if len(time) > 19 {
				time = time[:19]
			}
			fmt.Printf("%-20s %-8s %-15s %-15s %-10s %-10s %-15s %-20s %s\n",
				r.IP, r.Country, r.Region, r.City, r.Browser, r.OS, r.Path, truncateReferrer(r.Referrer), time)
		}
	}
}

func truncateReferrer(ref string) string {
	if len(ref) > 20 {
		return ref[:17] + "..."
	}
	if ref == "" {
		return "(direct)"
	}
	return ref
}
