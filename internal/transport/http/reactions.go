package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"forum/internal/service"
	ws "forum/internal/transport/websocket"
)

func (h *Handler) PostSubroutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/posts/"), "/")
	if rest == "" {
		http.NotFound(w, r)
		return
	}

	parts := strings.Split(rest, "/")
	postID, err := strconv.Atoi(parts[0])
	if err != nil || postID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный ID"})
		return
	}

	if len(parts) == 2 && parts[1] == "reaction" {
		h.postReaction(w, r, postID)
		return
	}

	if len(parts) == 1 {
		h.postByID(w, r, postID)
		return
	}

	http.NotFound(w, r)
}

type reactionRequest struct {
	Reaction string `json:"reaction"`
}

func (h *Handler) postReaction(w http.ResponseWriter, r *http.Request, postID int) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Только POST"})
		return
	}

	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	var req reactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный JSON"})
		return
	}

	update, err := h.Forum.SetReaction(userID, postID, req.Reaction)
	if err != nil {
		h.writeReactionError(w, err)
		return
	}

	h.Hub.BroadcastForum(ws.ReactionUpdatedMessage(update))
	writeJSON(w, http.StatusOK, update)
}

func (h *Handler) writeReactionError(w http.ResponseWriter, err error) {
	var ve service.ValidationError
	switch {
	case errors.As(err, &ve):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ve.Message})
	case errors.Is(err, service.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Пост не найден"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сервера"})
	}
}
