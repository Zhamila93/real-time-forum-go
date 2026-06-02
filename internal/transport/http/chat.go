package http

import (
	"errors"
	"net/http"
	"strconv"

	"forum/internal/service"
)

func (h *Handler) UsersList(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	users, err := h.Chat.ListUsers(currentUserID, h.Hub)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка БД"})
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	withStr := r.URL.Query().Get("with")
	withID, err := strconv.Atoi(withStr)
	if err != nil || withID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный параметр with"})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	msgs, err := h.Chat.GetMessages(currentUserID, withID, limit, offset)
	if err != nil {
		var ve service.ValidationError
		if errors.As(err, &ve) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": ve.Message})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка БД"})
		return
	}
	writeJSON(w, http.StatusOK, msgs)
}
