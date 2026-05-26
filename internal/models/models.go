package models

import "time"

type Site struct {
	ID        int64     `json:"id"`
	Domain    string    `json:"domain"`
	CreatedAt time.Time `json:"created_at"`
}

type IgnoredIP struct {
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

type Token struct {
	ID        int64     `json:"id"`
	Hash      string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenInfo struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type Pageview struct {
	ID           int64     `json:"id"`
	SiteID       int64     `json:"site_id"`
	Path         string    `json:"path"`
	Referrer     string    `json:"referrer"`
	Browser      string    `json:"browser"`
	OS           string    `json:"os"`
	ScreenWidth  int       `json:"screen_width"`
	ScreenHeight int       `json:"screen_height"`
	Country      string    `json:"country"`
	Region       string    `json:"region"`
	City         string    `json:"city"`
	Locality     string    `json:"locality"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
}

type PageviewInput struct {
	SiteID       int64
	Path         string
	Referrer     string
	Browser      string
	OS           string
	ScreenWidth  int
	ScreenHeight int
	Country      string
	Region       string
	City         string
	Locality     string
	IP           string
	UserAgent    string
}

type StatsRow struct {
	Site     string `json:"site"`
	IP       string `json:"ip"`
	Path     string `json:"path"`
	Count    int    `json:"count,omitempty"`
	Country  string `json:"country,omitempty"`
	Region   string `json:"region,omitempty"`
	City     string `json:"city,omitempty"`
	Browser  string `json:"browser,omitempty"`
	OS       string `json:"os,omitempty"`
	Referrer string `json:"referrer,omitempty"`
	Time     string `json:"time,omitempty"`
}
