package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"forum/backend/models"
)

// UsersList — список всех пользователей (кроме текущего), отсортирован:
// сначала те с кем была переписка (по дате последнего сообщения, новые сверху),
// затем остальные по алфавиту
func (h *Handler) UsersList(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	rows, err := h.DB.Query(`
		SELECT u.id, u.nickname,
		  (SELECT MAX(m.created_at) FROM messages m
		     WHERE (m.sender_id = u.id AND m.receiver_id = ?)
		        OR (m.sender_id = ? AND m.receiver_id = u.id)
		  ) AS last_msg
		FROM users u
		WHERE u.id != ?
	`, currentUserID, currentUserID, currentUserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка БД"})
		return
	}
	defer rows.Close()

	onlineMap := h.Hub.OnlineUsersMap()

	type item struct {
		ID          int        `json:"id"`
		Nickname    string     `json:"nickname"`
		Online      bool       `json:"online"`
		LastMessage *time.Time `json:"last_message"`
	}
	var withMsg []item
	var withoutMsg []item

	for rows.Next() {
		var it item
		var lastMsg *time.Time
		_ = rows.Scan(&it.ID, &it.Nickname, &lastMsg)
		it.LastMessage = lastMsg
		_, it.Online = onlineMap[it.ID]
		if lastMsg != nil {
			withMsg = append(withMsg, it)
		} else {
			withoutMsg = append(withoutMsg, it)
		}
	}

	// сортируем withMsg по убыванию даты (новые сверху)
	sort.Slice(withMsg, func(i, j int) bool {
		if withMsg[i].LastMessage == nil {
			return false
		}
		if withMsg[j].LastMessage == nil {
			return true
		}
		return withMsg[i].LastMessage.After(*withMsg[j].LastMessage)
	})
	// withoutMsg — по алфавиту
	sort.Slice(withoutMsg, func(i, j int) bool {
		return withoutMsg[i].Nickname < withoutMsg[j].Nickname
	})

	result := []models.UserListItem{}
	for _, x := range withMsg {
		result = append(result, models.UserListItem{
			ID: x.ID, Nickname: x.Nickname, Online: x.Online, LastMessage: x.LastMessage,
		})
	}
	for _, x := range withoutMsg {
		result = append(result, models.UserListItem{
			ID: x.ID, Nickname: x.Nickname, Online: x.Online, LastMessage: x.LastMessage,
		})
	}

	writeJSON(w, http.StatusOK, result)
}

// GetMessages — получить историю переписки с пользователем
// Параметры query:
//
//	?with=USER_ID  — собеседник
//	?offset=N      — смещение для пагинации
//	?limit=N       — сколько (по умолчанию 10)
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
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	// последние сообщения (от свежих к старым), затем перевернём
	rows, err := h.DB.Query(`
		SELECT m.id, m.sender_id, m.receiver_id,
		       us.nickname AS sender_nick, ur.nickname AS receiver_nick,
		       m.content, m.created_at
		FROM messages m
		JOIN users us ON us.id = m.sender_id
		JOIN users ur ON ur.id = m.receiver_id
		WHERE (m.sender_id = ? AND m.receiver_id = ?)
		   OR (m.sender_id = ? AND m.receiver_id = ?)
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, currentUserID, withID, withID, currentUserID, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка БД"})
		return
	}
	defer rows.Close()

	msgs := []models.Message{}
	for rows.Next() {
		var m models.Message
		_ = rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Sender, &m.Receiver, &m.Content, &m.CreatedAt)
		msgs = append(msgs, m)
	}
	// перевернём (в порядке возрастания времени)
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	writeJSON(w, http.StatusOK, msgs)
}
