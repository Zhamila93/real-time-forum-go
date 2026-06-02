package repository

import (
	"database/sql"
	"time"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) DeleteByUserID(userID int) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

func (r *SessionRepository) Create(sessionID string, userID int, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
		sessionID, userID, expiresAt,
	)
	return err
}

func (r *SessionRepository) Delete(sessionID string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
	return err
}

func (r *SessionRepository) Lookup(sessionID string) (userID int, expiresAt time.Time, err error) {
	err = r.db.QueryRow(
		`SELECT user_id, expires_at FROM sessions WHERE id = ?`,
		sessionID,
	).Scan(&userID, &expiresAt)
	return
}
