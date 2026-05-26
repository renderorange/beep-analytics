package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	conn.SetMaxOpenConns(1)

	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func migrate(conn *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS ignored_ips (
			ip TEXT PRIMARY KEY,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hash TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS pageviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			site_id INTEGER NOT NULL,
			path TEXT NOT NULL DEFAULT '',
			referrer TEXT NOT NULL DEFAULT '',
			browser TEXT NOT NULL DEFAULT '',
			os TEXT NOT NULL DEFAULT '',
			screen_width INTEGER NOT NULL DEFAULT 0,
			screen_height INTEGER NOT NULL DEFAULT 0,
			country TEXT NOT NULL DEFAULT '',
			region TEXT NOT NULL DEFAULT '',
			city TEXT NOT NULL DEFAULT '',
			locality TEXT NOT NULL DEFAULT '',
			ip TEXT NOT NULL DEFAULT '',
			user_agent TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (site_id) REFERENCES sites(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_pageviews_site ON pageviews(site_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pageviews_created ON pageviews(created_at)`,
	}

	for _, q := range queries {
		if _, err := conn.Exec(q); err != nil {
			return fmt.Errorf("exec migration: %w", err)
		}
	}
	return nil
}
