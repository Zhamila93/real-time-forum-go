package handlers

import (
	"net/http"

	ws "forum/backend/websocket"

	gws "github.com/gorilla/websocket"
)

var upgrader = gws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// одноpаgevый: разрешаем любые
		return true
	},
}

func (h *Handler) WSHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromRequest(r)
	if userID == 0 {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	var nickname string
	if err := h.DB.QueryRow(`SELECT nickname FROM users WHERE id = ?`, userID).Scan(&nickname); err != nil {
		http.Error(w, "Ошибка", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		Hub:      h.Hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userID,
		Nickname: nickname,
	}

	h.Hub.Register(client)
	go client.WritePump()
	go client.ReadPump()
}
