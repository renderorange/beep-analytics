package db

import (
	"fmt"
)

func (db *DB) AddIgnoredIP(ip string) error {
	_, err := db.conn.Exec("INSERT OR IGNORE INTO ignored_ips (ip) VALUES (?)", ip)
	if err != nil {
		return fmt.Errorf("insert ignored ip: %w", err)
	}
	return nil
}

func (db *DB) RemoveIgnoredIP(ip string) error {
	_, err := db.conn.Exec("DELETE FROM ignored_ips WHERE ip = ?", ip)
	return err
}

func (db *DB) ListIgnoredIPs() ([]string, error) {
	rows, err := db.conn.Query("SELECT ip FROM ignored_ips ORDER BY ip")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ips []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}

func (db *DB) IsIPIgnored(ip string) (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM ignored_ips WHERE ip = ?", ip).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// MatchesIgnoredIP checks if an IP matches any ignored IP or CIDR range.
// For simplicity, only exact matches are supported initially.
func (db *DB) MatchesIgnoredIP(ip string) (bool, error) {
	return db.IsIPIgnored(ip)
}
