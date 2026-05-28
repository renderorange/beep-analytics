package db

import (
	"fmt"

	"github.com/adventurehound/beep-analytics/internal/models"
)

func (db *DB) InsertPageview(pv models.PageviewInput) error {
	_, err := db.conn.Exec(
		`INSERT INTO pageviews (site_id, path, referrer, browser, os, screen_width, screen_height, country, region, city, locality, ip, user_agent)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		pv.SiteID, pv.Path, pv.Referrer, pv.Browser, pv.OS,
		pv.ScreenWidth, pv.ScreenHeight, pv.Country, pv.Region,
		pv.City, pv.Locality, pv.IP, pv.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("insert pageview: %w", err)
	}
	return nil
}
