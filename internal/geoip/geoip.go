package geoip

import (
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

type Info struct {
	Country  string
	Region   string
	City     string
	Locality string
}

type Lookup struct {
	enabled   bool
	blocks    []block
	locations map[int64]location
}

type block struct {
	network *net.IPNet
	geoID   int64
}

type location struct {
	Country string
	Region  string
	City    string
}

func New(mmdbPath string) (*Lookup, error) {
	if mmdbPath == "" {
		return &Lookup{enabled: false}, nil
	}

	l := &Lookup{enabled: true, locations: make(map[int64]location)}

	// Expect directory with GeoLite2-City-Blocks.csv and GeoLite2-City-Locations-en.csv
	dir := filepath.Dir(mmdbPath)

	if err := l.loadLocations(filepath.Join(dir, "GeoLite2-City-Locations-en.csv")); err != nil {
		return nil, fmt.Errorf("load locations: %w", err)
	}

	if err := l.loadBlocks(filepath.Join(dir, "GeoLite2-City-Blocks.csv")); err != nil {
		return nil, fmt.Errorf("load blocks: %w", err)
	}

	return l, nil
}

func (l *Lookup) loadLocations(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.LazyQuotes = true
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	for i, row := range records {
		if i == 0 {
			continue // header
		}
		if len(row) < 8 {
			continue
		}
		geoID, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			continue
		}
		l.locations[geoID] = location{
			Country: row[4],
			Region:  row[5],
			City:    row[7],
		}
	}
	return nil
}

func (l *Lookup) loadBlocks(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.LazyQuotes = true
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	for i, row := range records {
		if i == 0 {
			continue // header
		}
		if len(row) < 2 {
			continue
		}
		_, network, err := net.ParseCIDR(row[0])
		if err != nil {
			continue
		}
		geoID, err := strconv.ParseInt(row[1], 10, 64)
		if err != nil {
			continue
		}
		l.blocks = append(l.blocks, block{network: network, geoID: geoID})
	}
	return nil
}

func (l *Lookup) LookupIP(ipStr string) Info {
	if !l.enabled {
		return Info{}
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return Info{}
	}

	// Find matching block
	for _, b := range l.blocks {
		if b.network.Contains(ip) {
			if loc, ok := l.locations[b.geoID]; ok {
				return Info{
					Country:  loc.Country,
					Region:   loc.Region,
					City:     loc.City,
					Locality: "",
				}
			}
		}
	}

	return Info{}
}

func (l *Lookup) Enabled() bool {
	return l.enabled
}
