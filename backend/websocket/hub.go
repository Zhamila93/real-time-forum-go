package websocket

import (
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Сообщение, передаваемое по WS
type WSMessage struct {
	Type      string `json:"type"` // "message", "typing", "user_online", "user_offline", "online_list", "new_message"
	From      int    `json:"from,omitempty"`
	FromNick  string `json:"from_nick,omitempty"`
	To        int    `json:"to,omitempty"`
	Content   string `json:"content,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	OnlineIDs []int  `json:"online_ids,omitempty"`
}

// Client — клиент WS
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   int
	Nickname string
}

// Hub — центральная точка управления WS-соединениями
type Hub struct {
	clients    map[*Client]bool
	usersConns map[int]map[*Client]bool // userID -> set клиентов
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
	db         *sql.DB
}

func NewHub(db *sql.DB) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		usersConns: make(map[int]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		db:         db,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = true
			if h.usersConns[c.UserID] == nil {
				h.usersConns[c.UserID] = make(map[*Client]bool)
			}
			wasFirstConn := len(h.usersConns[c.UserID]) == 0
			h.usersConns[c.UserID][c] = true
			h.mu.Unlock()

			// уведомляем всех, что пользователь онлайн (только при первом подключении)
			if wasFirstConn {
				h.broadcastUserStatus(c.UserID, true)
			}
			// отправляем новому клиенту список всех онлайн
			h.sendOnlineList(c)

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				if conns, ok := h.usersConns[c.UserID]; ok {
					delete(conns, c)
					if len(conns) == 0 {
						delete(h.usersConns, c.UserID)
						// уведомим всех, что пользователь оффлайн
						go h.broadcastUserStatus(c.UserID, false)
					}
				}
				close(c.Send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.Send <- msg:
				default:
					// канал переполнен — пропускаем
				}
			}
			h.mu.RUnlock()
		}
	}
}

// OnlineUsersMap — карта userID -> true для онлайн-юзеров
func (h *Hub) OnlineUsersMap() map[int]bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	res := make(map[int]bool, len(h.usersConns))
	for uid := range h.usersConns {
		res[uid] = true
	}
	return res
}

func (h *Hub) onlineIDs() []int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]int, 0, len(h.usersConns))
	for uid := range h.usersConns {
		ids = append(ids, uid)
	}
	return ids
}

func (h *Hub) sendOnlineList(c *Client) {
	msg := WSMessage{
		Type:      "online_list",
		OnlineIDs: h.onlineIDs(),
	}
	data, _ := json.Marshal(msg)
	select {
	case c.Send <- data:
	default:
	}
}

func (h *Hub) broadcastUserStatus(userID int, online bool) {
	t := "user_online"
	if !online {
		t = "user_offline"
	}
	msg := WSMessage{
		Type: t,
		From: userID,
	}
	data, _ := json.Marshal(msg)
	h.broadcast <- data
}

// SendToUser — отправить сообщение конкретному пользователю на все его соединения
func (h *Hub) SendToUser(userID int, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns, ok := h.usersConns[userID]
	if !ok {
		return
	}
	for c := range conns {
		select {
		case c.Send <- data:
		default:
		}
	}
}

// Register / Unregister — экспорт для handlers
func (h *Hub) Register(c *Client) {
	h.register <- c
}
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

// HandleIncoming — обрабатывает входящие сообщения от клиента
func (h *Hub) HandleIncoming(c *Client, raw []byte) {
	var msg WSMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Println("ws: некорректный JSON от клиента:", err)
		return
	}

	switch msg.Type {
	case "message":
		h.handleChatMessage(c, msg)
	case "typing":
		// можно реализовать индикатор «печатает»
		out := WSMessage{
			Type:     "typing",
			From:     c.UserID,
			FromNick: c.Nickname,
			To:       msg.To,
		}
		data, _ := json.Marshal(out)
		h.SendToUser(msg.To, data)
	}
}

func (h *Hub) handleChatMessage(c *Client, msg WSMessage) {
	if msg.To == 0 || msg.Content == "" {
		return
	}

	// сохраняем в БД
	res, err := h.db.Exec(
		`INSERT INTO messages (sender_id, receiver_id, content) VALUES (?, ?, ?)`,
		c.UserID, msg.To, msg.Content)
	if err != nil {
		log.Println("ws: ошибка сохранения сообщения:", err)
		return
	}
	id, _ := res.LastInsertId()
	now := time.Now().UTC().Format(time.RFC3339)

	payload, _ := json.Marshal(map[string]interface{}{
		"type":       "new_message",
		"id":         id,
		"from":       c.UserID,
		"from_nick":  c.Nickname,
		"to":         msg.To,
		"content":    msg.Content,
		"created_at": now,
	})

	// отправляем получателю и отправителю (для синхронизации других вкладок)
	h.SendToUser(msg.To, payload)
	h.SendToUser(c.UserID, payload)
}
