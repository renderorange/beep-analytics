package geoip

import (
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Info struct {
	Country  string
	Region   string
	City     string
	Locality string
}

type Lookup struct {
	enabled   bool
	mu        sync.RWMutex
	blocks    []block
	locations map[int64]location
	ready     bool
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

	// Verify the path exists before starting async load
	dir := mmdbPath
	if info, err := os.Stat(mmdbPath); err != nil || !info.IsDir() {
		dir = filepath.Dir(mmdbPath)
	}
	locPath := filepath.Join(dir, "GeoLite2-City-Locations-en.csv")
	blkPath := filepath.Join(dir, "GeoLite2-City-Blocks.csv")
	if _, err := os.Stat(locPath); err != nil {
		return nil, fmt.Errorf("load locations: %w", err)
	}
	if _, err := os.Stat(blkPath); err != nil {
		return nil, fmt.Errorf("load blocks: %w", err)
	}

	l := &Lookup{
		enabled:   true,
		locations: make(map[int64]location),
	}

	go l.loadAsync(locPath, blkPath)

	return l, nil
}

func (l *Lookup) loadAsync(locPath, blkPath string) {
	log.Println("GeoIP: loading data...")

	locations := make(map[int64]location)
	if err := loadLocations(locPath, locations); err != nil {
		log.Printf("GeoIP: error loading locations: %v", err)
		return
	}

	var blocks []block
	if err := loadBlocks(blkPath, &blocks); err != nil {
		log.Printf("GeoIP: error loading blocks: %v", err)
		return
	}

	l.mu.Lock()
	l.locations = locations
	l.blocks = blocks
	l.ready = true
	l.mu.Unlock()

	log.Printf("GeoIP: loaded %d locations, %d blocks", len(locations), len(blocks))
}

func loadLocations(path string, locations map[int64]location) error {
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
		if len(row) < 11 {
			continue
		}
		geoID, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			continue
		}
		locations[geoID] = location{
			Country: row[4],
			Region:  row[7],
			City:    row[10],
		}
	}
	return nil
}

func loadBlocks(path string, blocks *[]block) error {
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
		*blocks = append(*blocks, block{network: network, geoID: geoID})
	}
	return nil
}

func (l *Lookup) LookupIP(ipStr string) Info {
	if !l.enabled {
		return Info{}
	}

	l.mu.RLock()
	ready := l.ready
	blocks := l.blocks
	locations := l.locations
	l.mu.RUnlock()

	if !ready {
		return Info{}
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return Info{}
	}

	// Find matching block
	for _, b := range blocks {
		if b.network.Contains(ip) {
			if loc, ok := locations[b.geoID]; ok {
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

func (l *Lookup) Ready() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.ready
}
