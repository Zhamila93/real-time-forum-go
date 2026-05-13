package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"forum/backend/models"
)

// Categories — получить список категорий
func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT id, name FROM categories ORDER BY id`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка БД"})
		return
	}
	defer rows.Close()

	type cat struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	var cats []cat
	for rows.Next() {
		var c cat
		_ = rows.Scan(&c.ID, &c.Name)
		cats = append(cats, c)
	}
	writeJSON(w, http.StatusOK, cats)
}

// Posts — GET список всех постов, POST создать новый пост
func (h *Handler) Posts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listPosts(w, r)
	case http.MethodPost:
		h.createPost(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не разрешён"})
	}
}

func (h *Handler) listPosts(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}

	category := r.URL.Query().Get("category")

	query := `
		SELECT p.id, p.user_id, u.nickname, p.title, p.content, p.created_at,
		       (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id) AS cnt
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

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка БД"})
		return
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.Title, &p.Content, &p.CreatedAt, &p.CommentCount); err != nil {
			continue
		}
		p.Categories = h.getPostCategories(p.ID)
		posts = append(posts, p)
	}
	writeJSON(w, http.StatusOK, posts)
}

func (h *Handler) getPostCategories(postID int) []string {
	rows, err := h.DB.Query(`
		SELECT c.name FROM categories c
		JOIN post_categories pc ON pc.category_id = c.id
		WHERE pc.post_id = ?`, postID)
	if err != nil {
		return []string{}
	}
	defer rows.Close()
	cats := []string{}
	for rows.Next() {
		var name string
		_ = rows.Scan(&name)
		cats = append(cats, name)
	}
	return cats
}

type createPostRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Categories []string `json:"categories"`
}

func (h *Handler) createPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный JSON"})
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	if req.Title == "" || req.Content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Заполните заголовок и текст"})
		return
	}
	if len(req.Categories) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Выберите хотя бы одну категорию"})
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка БД"})
		return
	}

	res, err := tx.Exec(`INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)`,
		userID, req.Title, req.Content)
	if err != nil {
		tx.Rollback()
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сохранения"})
		return
	}
	postID, _ := res.LastInsertId()

	for _, catName := range req.Categories {
		var catID int
		if err := tx.QueryRow(`SELECT id FROM categories WHERE name = ?`, catName).Scan(&catID); err != nil {
			continue
		}
		_, _ = tx.Exec(`INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES (?, ?)`,
			postID, catID)
	}

	if err := tx.Commit(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка коммита"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"id": postID})
}

// PostByID — получить пост с комментариями
func (h *Handler) PostByID(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный ID"})
		return
	}

	var p models.Post
	err = h.DB.QueryRow(`
		SELECT p.id, p.user_id, u.nickname, p.title, p.content, p.created_at
		FROM posts p JOIN users u ON u.id = p.user_id
		WHERE p.id = ?`, id).Scan(&p.ID, &p.UserID, &p.Nickname, &p.Title, &p.Content, &p.CreatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Пост не найден"})
		return
	}
	p.Categories = h.getPostCategories(p.ID)

	// комментарии
	rows, err := h.DB.Query(`
		SELECT c.id, c.post_id, c.user_id, u.nickname, c.content, c.created_at
		FROM comments c JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ? ORDER BY c.created_at ASC`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка БД"})
		return
	}
	defer rows.Close()

	comments := []models.Comment{}
	for rows.Next() {
		var c models.Comment
		_ = rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Nickname, &c.Content, &c.CreatedAt)
		comments = append(comments, c)
	}
	p.CommentCount = len(comments)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"post":     p,
		"comments": comments,
	})
}

type createCommentRequest struct {
	PostID  int    `json:"post_id"`
	Content string `json:"content"`
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Только POST"})
		return
	}
	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный JSON"})
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" || req.PostID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Введите комментарий"})
		return
	}

	res, err := h.DB.Exec(`INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)`,
		req.PostID, userID, req.Content)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сохранения"})
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusOK, map[string]interface{}{"id": id})
}
