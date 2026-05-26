package db

import (
	"database/sql"
	"fmt"

	"github.com/adventurehound/beep/internal/models"
)

func (db *DB) AddSite(domain string) (*models.Site, error) {
	result, err := db.conn.Exec("INSERT INTO sites (domain) VALUES (?)", domain)
	if err != nil {
		return nil, fmt.Errorf("insert site: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}
	return &models.Site{ID: id, Domain: domain}, nil
}

func (db *DB) RemoveSite(domain string) error {
	_, err := db.conn.Exec("DELETE FROM sites WHERE domain = ?", domain)
	return err
}

func (db *DB) ListSites() ([]models.Site, error) {
	rows, err := db.conn.Query("SELECT id, domain, created_at FROM sites ORDER BY domain")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []models.Site
	for rows.Next() {
		var s models.Site
		if err := rows.Scan(&s.ID, &s.Domain, &s.CreatedAt); err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}
	return sites, rows.Err()
}

func (db *DB) GetSiteByDomain(domain string) (*models.Site, error) {
	var s models.Site
	err := db.conn.QueryRow("SELECT id, domain, created_at FROM sites WHERE domain = ?", domain).
		Scan(&s.ID, &s.Domain, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (db *DB) GetSiteByID(id int64) (*models.Site, error) {
	var s models.Site
	err := db.conn.QueryRow("SELECT id, domain, created_at FROM sites WHERE id = ?", id).
		Scan(&s.ID, &s.Domain, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
