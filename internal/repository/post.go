package repository

import (
	"database/sql"

	"forum/internal/domain"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) List(category string) ([]domain.Post, error) {
	query := `
		SELECT p.id, p.user_id, u.nickname, p.title, p.content, p.created_at,
		       (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id) AS cnt,
		       (SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id AND r.value = 1) AS likes,
		       (SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id AND r.value = -1) AS dislikes
		FROM posts p
		JOIN users u ON u.id = p.user_id
	`
	args := []interface{}{}
	if category != "" {
		query += `
			WHERE p.id IN (
				SELECT pc.post_id FROM post_categories pc
				JOIN categories c ON c.id = pc.category_id
				WHERE c.name = ?
			)
		`
		args = append(args, category)
	}
	query += ` ORDER BY p.created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []domain.Post
	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.Title, &p.Content, &p.CreatedAt, &p.CommentCount, &p.LikeCount, &p.DislikeCount); err != nil {
			continue
		}
		p.Categories, _ = r.CategoriesForPost(p.ID)
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *PostRepository) CategoriesForPost(postID int) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT c.name FROM categories c
		JOIN post_categories pc ON pc.category_id = c.id
		WHERE pc.post_id = ?`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		cats = append(cats, name)
	}
	return cats, rows.Err()
}

func (r *PostRepository) GetByID(id int) (domain.Post, error) {
	var p domain.Post
	err := r.db.QueryRow(`
		SELECT p.id, p.user_id, u.nickname, p.title, p.content, p.created_at,
		       (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id),
		       (SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id AND r.value = 1),
		       (SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id AND r.value = -1)
		FROM posts p JOIN users u ON u.id = p.user_id
		WHERE p.id = ?`, id).Scan(
		&p.ID, &p.UserID, &p.Nickname, &p.Title, &p.Content, &p.CreatedAt,
		&p.CommentCount, &p.LikeCount, &p.DislikeCount,
	)
	if err != nil {
		return p, err
	}
	p.Categories, _ = r.CategoriesForPost(p.ID)
	return p, nil
}

func (r *PostRepository) Create(userID int, title, content string, categoryNames []string) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)`,
		userID, title, content)
	if err != nil {
		return 0, err
	}
	postID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, catName := range categoryNames {
		var catID int
		if err := tx.QueryRow(`SELECT id FROM categories WHERE name = ?`, catName).Scan(&catID); err != nil {
			continue
		}
		_, _ = tx.Exec(`INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES (?, ?)`,
			postID, catID)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return postID, nil
}
