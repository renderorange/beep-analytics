package db

import (
	"fmt"
	"time"

	"github.com/adventurehound/beep/internal/models"
)

type StatsQuery struct {
	SiteID int64
	From   time.Time
	To     time.Time
}

func (db *DB) GetPageviewsBySite(siteID int64) (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM pageviews WHERE site_id = ?", siteID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count pageviews: %w", err)
	}
	return count, nil
}

func (db *DB) GetAggregateStats(q StatsQuery) ([]models.StatsRow, error) {
	query := `SELECT s.domain, pv.ip, pv.path, COUNT(*) as count
	          FROM pageviews pv
	          JOIN sites s ON pv.site_id = s.id
	          WHERE pv.created_at >= ? AND pv.created_at <= ?`
	args := []interface{}{q.From.Format(time.RFC3339), q.To.Format(time.RFC3339)}

	if q.SiteID > 0 {
		query += " AND pv.site_id = ?"
		args = append(args, q.SiteID)
	}

	query += " GROUP BY s.domain, pv.ip, pv.path ORDER BY s.domain, count DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query stats: %w", err)
	}
	defer rows.Close()

	var results []models.StatsRow
	for rows.Next() {
		var r models.StatsRow
		if err := rows.Scan(&r.Site, &r.IP, &r.Path, &r.Count); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (db *DB) GetVerboseStats(q StatsQuery) ([]models.StatsRow, error) {
	query := `SELECT s.domain, pv.ip, pv.country, pv.browser, pv.os, pv.path, pv.referrer, pv.created_at
	          FROM pageviews pv
	          JOIN sites s ON pv.site_id = s.id
	          WHERE pv.created_at >= ? AND pv.created_at <= ?`
	args := []interface{}{q.From.Format(time.RFC3339), q.To.Format(time.RFC3339)}

	if q.SiteID > 0 {
		query += " AND pv.site_id = ?"
		args = append(args, q.SiteID)
	}

	query += " ORDER BY s.domain, pv.created_at DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query verbose stats: %w", err)
	}
	defer rows.Close()

	var results []models.StatsRow
	for rows.Next() {
		var r models.StatsRow
		if err := rows.Scan(&r.Site, &r.IP, &r.Country, &r.Browser, &r.OS, &r.Path, &r.Referrer, &r.Time); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
