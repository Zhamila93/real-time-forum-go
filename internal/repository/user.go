package repository

import (
	"database/sql"
	"time"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) ExistsByNicknameOrEmail(nickname, email string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM users WHERE nickname = ? OR email = ?`,
		nickname, email,
	).Scan(&count)
	return count > 0, err
}

func (r *UserRepository) Create(nickname string, age int, gender, firstName, lastName, email, passwordHash string) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO users (nickname, age, gender, first_name, last_name, email, password)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		nickname, age, gender, firstName, lastName, email, passwordHash,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *UserRepository) FindByIdentifier(nicknameOrEmail, emailLower string) (id int, nickname, passwordHash string, err error) {
	err = r.db.QueryRow(
		`SELECT id, nickname, password FROM users WHERE nickname = ? OR email = ?`,
		nicknameOrEmail, emailLower,
	).Scan(&id, &nickname, &passwordHash)
	return
}

func (r *UserRepository) NicknameByID(userID int) (string, error) {
	var nickname string
	err := r.db.QueryRow(`SELECT nickname FROM users WHERE id = ?`, userID).Scan(&nickname)
	return nickname, err
}

type UserRow struct {
	ID          int
	Nickname    string
	LastMessage *time.Time
}

func (r *UserRepository) ListExcept(currentUserID int) ([]UserRow, error) {
	rows, err := r.db.Query(`
		SELECT u.id, u.nickname,
		  (SELECT MAX(m.created_at) FROM messages m
		     WHERE (m.sender_id = u.id AND m.receiver_id = ?)
		        OR (m.sender_id = ? AND m.receiver_id = u.id)
		  ) AS last_msg
		FROM users u
		WHERE u.id != ?
	`, currentUserID, currentUserID, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []UserRow
	for rows.Next() {
		var row UserRow
		var lastMsg sql.NullTime
		if err := rows.Scan(&row.ID, &row.Nickname, &lastMsg); err != nil {
			continue
		}
		if lastMsg.Valid {
			t := lastMsg.Time
			row.LastMessage = &t
		}
		list = append(list, row)
	}
	return list, rows.Err()
}
