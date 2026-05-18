package websocket

import (
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"

	gws "github.com/gorilla/websocket"
)

type WSMessage struct {
	Type      string `json:"type"`
	ID        int    `json:"id,omitempty"`
	From      int    `json:"from,omitempty"`
	FromNick  string `json:"from_nick,omitempty"`
	To        int    `json:"to,omitempty"`
	Content   string `json:"content,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	OnlineIDs []int  `json:"online_ids,omitempty"`
}

type Client struct {
	Hub      *Hub
	Conn     *gws.Conn
	Send     chan []byte
	UserID   int
	Nickname string
}

type Hub struct {
	clients    map[*Client]bool
	usersConns map[int]map[*Client]bool
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
			
			isFirst := len(h.usersConns[c.UserID]) == 0
			h.usersConns[c.UserID][c] = true
			
			// Получаем список ID под замком
			ids := h.getOnlineIDsLocked()
			h.mu.Unlock()

			// 1. Отправляем новичку список всех, кто онлайн
			h.sendDirect(c, WSMessage{
				Type:      "online_list",
				OnlineIDs: ids,
			})

			// 2. Если это первое зашедшее устройство юзера, уведомляем остальных
			if isFirst {
				h.broadcastStatus(c.UserID, true)
			}

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				
				conns := h.usersConns[c.UserID]
				delete(conns, c)
				
				isLast := len(conns) == 0
				if isLast {
					delete(h.usersConns, c.UserID)
				}
				
				close(c.Send)
				h.mu.Unlock()

				if isLast {
					h.broadcastStatus(c.UserID, false)
				}
			} else {
				h.mu.Unlock()
			}

		case msg := <-h.broadcast:
			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.Send <- msg:
				default:
					// Если клиент не успевает читать — не вешаем сервер
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Вспомогательные методы (без блокировок внутри, вызываются под Lock)
func (h *Hub) getOnlineIDsLocked() []int {
	ids := make([]int, 0, len(h.usersConns))
	for uid := range h.usersConns {
		ids = append(ids, uid)
	}
	return ids
}

// Публичные методы для рассылки (безопасные)
func (h *Hub) broadcastStatus(userID int, online bool) {
	msgType := "user_offline"
	if online {
		msgType = "user_online"
	}
	
	msg := WSMessage{Type: msgType, From: userID}
	data, _ := json.Marshal(msg)
	h.broadcast <- data
}

func (h *Hub) sendDirect(c *Client, msg WSMessage) {
	data, _ := json.Marshal(msg)
	select {
	case c.Send <- data:
	default:
	}
}

func (h *Hub) SendToUser(userID int, msg WSMessage) {
	data, _ := json.Marshal(msg)
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if conns, ok := h.usersConns[userID]; ok {
		for c := range conns {
			select {
			case c.Send <- data:
			default:
			}
		}
	}
}

func (h *Hub) HandleIncoming(c *Client, raw []byte) {
	var msg WSMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("ws: bad json from user %d: %v raw=%s", c.UserID, err, string(raw))
		return
	}

	switch msg.Type {
	case "message":
		if msg.To == 0 || msg.Content == "" {
			log.Printf("ws: skip message from %d: to=%d content_len=%d", c.UserID, msg.To, len(msg.Content))
			return
		}

		// Сохранение в БД
		res, err := h.db.Exec(
			`INSERT INTO messages (sender_id, receiver_id, content) VALUES (?, ?, ?)`,
			c.UserID, msg.To, msg.Content,
		)
		if err != nil {
			log.Println("DB Error:", err)
			return
		}

		lastID, err := res.LastInsertId()
		if err != nil {
			log.Println("LastInsertId:", err)
		}

		createdAt := time.Now().Format(time.RFC3339)

		response := WSMessage{
			Type:      "new_message",
			ID:        int(lastID),
			From:      c.UserID,
			FromNick:  c.Nickname,
			To:        msg.To,
			Content:   msg.Content,
			CreatedAt: createdAt,
		}

		// Отправляем обоим сторонам (отправитель и получатель видят без перезагрузки)
		h.SendToUser(msg.To, response)
		h.SendToUser(c.UserID, response)

	case "typing":
		if msg.To == 0 {
			return
		}
		h.SendToUser(msg.To, WSMessage{
			Type:     "typing",
			From:     c.UserID,
			FromNick: c.Nickname,
			To:       msg.To,
		})
	}
}

func (h *Hub) Register(c *Client)   { h.register <- c }
func (h *Hub) Unregister(c *Client) { h.unregister <- c }

func (h *Hub) OnlineUsersMap() map[int]bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	res := make(map[int]bool, len(h.usersConns))
	for uid := range h.usersConns {
		res[uid] = true
	}
	return res
}