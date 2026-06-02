package repository

import (
	"database/sql"

	"forum/internal/domain"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) ListByPostID(postID int) ([]domain.Comment, error) {
	rows, err := r.db.Query(`
		SELECT c.id, c.post_id, c.user_id, u.nickname, c.content, c.created_at
		FROM comments c JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ? ORDER BY c.created_at ASC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Nickname, &c.Content, &c.CreatedAt); err != nil {
			continue
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (r *CommentRepository) Create(postID, userID int, content string) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)`,
		postID, userID, content,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CommentRepository) GetByID(id int) (domain.Comment, error) {
	var c domain.Comment
	err := r.db.QueryRow(`
		SELECT c.id, c.post_id, c.user_id, u.nickname, c.content, c.created_at
		FROM comments c JOIN users u ON u.id = c.user_id
		WHERE c.id = ?`, id).Scan(
		&c.ID, &c.PostID, &c.UserID, &c.Nickname, &c.Content, &c.CreatedAt,
	)
	return c, err
}
