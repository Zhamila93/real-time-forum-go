package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"forum/internal/service"
)

type registerRequest struct {
	Nickname  string `json:"nickname"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Только POST"})
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный JSON"})
		return
	}

	result, err := h.Auth.Register(service.RegisterInput{
		Nickname:  req.Nickname,
		Age:       req.Age,
		Gender:    req.Gender,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	})
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	setSessionCookie(w, result.Session)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":       result.UserID,
		"nickname": result.Nickname,
	})
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Только POST"})
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный JSON"})
		return
	}

	result, err := h.Auth.Login(service.LoginInput{
		Identifier: req.Identifier,
		Password:   req.Password,
	})
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	setSessionCookie(w, result.Session)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":       result.UserID,
		"nickname": result.Nickname,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		_ = h.Auth.Logout(cookie.Value)
	}
	clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := h.userIDFromRequest(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Не авторизован"})
		return
	}
	id, nickname, err := h.Auth.Me(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":       id,
		"nickname": nickname,
	})
}

func (h *Handler) writeAuthError(w http.ResponseWriter, err error) {
	var ve service.ValidationError
	switch {
	case errors.As(err, &ve):
		status := http.StatusBadRequest
		if strings.Contains(ve.Message, "заняты") {
			status = http.StatusConflict
		}
		writeJSON(w, status, map[string]string{"error": ve.Message})
	case errors.Is(err, service.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Неверный логин или пароль"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сервера"})
	}
}
