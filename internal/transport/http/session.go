package http

import (
	"net/http"

	"forum/internal/service"
)

const SessionCookieName = "forum_session"

func setSessionCookie(w http.ResponseWriter, sess service.SessionInfo) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sess.ID,
		Path:     "/",
		Expires:  sess.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func (h *Handler) userIDFromRequest(r *http.Request) int {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return 0
	}
	return h.Auth.UserIDFromSession(cookie.Value)
}

func (h *Handler) requireAuth(w http.ResponseWriter, r *http.Request) (int, bool) {
	userID := h.userIDFromRequest(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Не авторизован"})
		return 0, false
	}
	return userID, true
}
