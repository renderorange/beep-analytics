package db

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/adventurehound/beep/internal/models"
)

func (db *DB) GenerateToken() (string, int64, error) {
	// Generate random token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", 0, fmt.Errorf("generate random: %w", err)
	}
	token := hex.EncodeToString(b)

	// Hash it for storage
	hash := hashToken(token)

	result, err := db.conn.Exec("INSERT INTO tokens (hash) VALUES (?)", hash)
	if err != nil {
		return "", 0, fmt.Errorf("insert token: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return "", 0, fmt.Errorf("get last insert id: %w", err)
	}

	return token, id, nil
}

func (db *DB) RevokeToken(id int64) error {
	_, err := db.conn.Exec("DELETE FROM tokens WHERE id = ?", id)
	return err
}

func (db *DB) ValidateToken(token string) (bool, error) {
	hash := hashToken(token)
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM tokens WHERE hash = ?", hash).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *DB) HasAnyToken() (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM tokens").Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *DB) ListTokens() ([]models.TokenInfo, error) {
	rows, err := db.conn.Query("SELECT id, created_at FROM tokens ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []models.TokenInfo
	for rows.Next() {
		var t models.TokenInfo
		if err := rows.Scan(&t.ID, &t.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
