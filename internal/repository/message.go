package repository

import (
	"database/sql"

	"forum/internal/domain"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(senderID, receiverID int, content string) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO messages (sender_id, receiver_id, content) VALUES (?, ?, ?)`,
		senderID, receiverID, content,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *MessageRepository) ListBetween(userID, withID, limit, offset int) ([]domain.Message, error) {
	rows, err := r.db.Query(`
		SELECT m.id, m.sender_id, m.receiver_id,
		       us.nickname AS sender_nick, ur.nickname AS receiver_nick,
		       m.content, m.created_at
		FROM messages m
		JOIN users us ON us.id = m.sender_id
		JOIN users ur ON ur.id = m.receiver_id
		WHERE (m.sender_id = ? AND m.receiver_id = ?)
		   OR (m.sender_id = ? AND m.receiver_id = ?)
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, userID, withID, withID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Sender, &m.Receiver, &m.Content, &m.CreatedAt); err != nil {
			continue
		}
		msgs = append(msgs, m)
	}
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, rows.Err()
}
