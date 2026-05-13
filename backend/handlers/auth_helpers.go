package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const sessionCookieName = "forum_session"
const sessionDuration = 24 * time.Hour * 7 // неделя

// CreateSession создаёт новую сессию для пользователя и сохраняет её в БД
func (h *Handler) createSession(w http.ResponseWriter, userID int) error {
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(sessionDuration)

	// удалим старые сессии этого пользователя — пользователь может иметь только одну активную
	_, _ = h.DB.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)

	_, err := h.DB.Exec(`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
		sessionID, userID, expiresAt)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// getUserIDFromRequest возвращает ID пользователя по cookie, либо 0
func (h *Handler) getUserIDFromRequest(r *http.Request) int {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return 0
	}
	var userID int
	var expiresAt time.Time
	err = h.DB.QueryRow(`SELECT user_id, expires_at FROM sessions WHERE id = ?`, cookie.Value).
		Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		return 0
	}
	if time.Now().After(expiresAt) {
		_, _ = h.DB.Exec(`DELETE FROM sessions WHERE id = ?`, cookie.Value)
		return 0
	}
	return userID
}

// requireAuth — проверка авторизации в обработчиках
func (h *Handler) requireAuth(w http.ResponseWriter, r *http.Request) (int, bool) {
	userID := h.getUserIDFromRequest(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Не авторизован"})
		return 0, false
	}
	return userID, true
}

// writeJSON — утилита для ответа JSON
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
