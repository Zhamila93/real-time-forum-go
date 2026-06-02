package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"forum/internal/service"
)

func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.Forum.ListCategories()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка БД"})
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

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
	posts, err := h.Forum.ListPosts(r.URL.Query().Get("category"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка БД"})
		return
	}
	writeJSON(w, http.StatusOK, posts)
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

	post, err := h.Forum.CreatePost(userID, service.CreatePostInput{
		Title:      req.Title,
		Content:    req.Content,
		Categories: req.Categories,
	})
	if err != nil {
		h.writeForumError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, post)
}

func (h *Handler) postByID(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Только GET"})
		return
	}

	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	post, comments, err := h.Forum.GetPostWithComments(id, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Пост не найден"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка БД"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"post":     post,
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

	comment, err := h.Forum.CreateComment(userID, req.PostID, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Пост не найден"})
			return
		}
		h.writeForumError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, comment)
}

func (h *Handler) writeForumError(w http.ResponseWriter, err error) {
	var ve service.ValidationError
	if errors.As(err, &ve) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ve.Message})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сервера"})
}
