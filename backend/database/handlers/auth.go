package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
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

	// валидация
	req.Nickname = strings.TrimSpace(req.Nickname)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Gender = strings.TrimSpace(req.Gender)

	if req.Nickname == "" || req.Email == "" || req.Password == "" ||
		req.FirstName == "" || req.LastName == "" || req.Gender == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Все поля обязательны"})
		return
	}
	if req.Age < 1 || req.Age > 150 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный возраст"})
		return
	}
	if len(req.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Пароль минимум 6 символов"})
		return
	}
	if !strings.Contains(req.Email, "@") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный e-mail"})
		return
	}

	// уникальность
	var exists int
	_ = h.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE nickname = ? OR email = ?`,
		req.Nickname, req.Email).Scan(&exists)
	if exists > 0 {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Никнейм или e-mail уже заняты"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка хеширования"})
		return
	}

	res, err := h.DB.Exec(`INSERT INTO users (nickname, age, gender, first_name, last_name, email, password)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		req.Nickname, req.Age, req.Gender, req.FirstName, req.LastName, req.Email, string(hash))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сохранения"})
		return
	}
	userID, _ := res.LastInsertId()

	if err := h.createSession(w, int(userID)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сессии"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":       userID,
		"nickname": req.Nickname,
	})
}

type loginRequest struct {
	Identifier string `json:"identifier"` // nickname или email
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
	req.Identifier = strings.TrimSpace(req.Identifier)
	if req.Identifier == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Введите логин и пароль"})
		return
	}

	var (
		userID   int
		nickname string
		hash     string
	)
	err := h.DB.QueryRow(`SELECT id, nickname, password FROM users
		WHERE nickname = ? OR email = ?`,
		req.Identifier, strings.ToLower(req.Identifier)).Scan(&userID, &nickname, &hash)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Неверный логин или пароль"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка БД"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Неверный логин или пароль"})
		return
	}

	if err := h.createSession(w, userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сессии"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":       userID,
		"nickname": nickname,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		_, _ = h.DB.Exec(`DELETE FROM sessions WHERE id = ?`, cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Me — возвращает текущего пользователя по сессии
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromRequest(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Не авторизован"})
		return
	}
	var nickname string
	if err := h.DB.QueryRow(`SELECT nickname FROM users WHERE id = ?`, userID).Scan(&nickname); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":       userID,
		"nickname": nickname,
	})
}
