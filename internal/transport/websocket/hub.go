package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"forum/internal/domain"
	"forum/internal/service"
)

// Messenger persists and validates outbound chat payloads.
type Messenger interface {
	SendPrivateMessage(senderID int, senderNick string, to int, content string) (domain.ChatOutbound, error)
}

type Hub struct {
	clients    map[*Client]bool
	usersConns map[int]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
	chat       Messenger
}

func NewHub(chat Messenger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		usersConns: make(map[int]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		chat:       chat,
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
			ids := h.getOnlineIDsLocked()
			h.mu.Unlock()

			h.sendDirect(c, Message{Type: "online_list", OnlineIDs: ids})
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
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) getOnlineIDsLocked() []int {
	ids := make([]int, 0, len(h.usersConns))
	for uid := range h.usersConns {
		ids = append(ids, uid)
	}
	return ids
}

func (h *Hub) broadcastStatus(userID int, online bool) {
	msgType := "user_offline"
	if online {
		msgType = "user_online"
	}
	data, _ := json.Marshal(Message{Type: msgType, From: userID})
	h.broadcast <- data
}

func (h *Hub) sendDirect(c *Client, msg Message) {
	data, _ := json.Marshal(msg)
	select {
	case c.Send <- data:
	default:
	}
}

func (h *Hub) SendToUser(userID int, msg Message) {
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
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("ws: bad json from user %d: %v", c.UserID, err)
		return
	}

	switch msg.Type {
	case "message":
		if msg.To == 0 || msg.Content == "" {
			return
		}
		out, err := h.chat.SendPrivateMessage(c.UserID, c.Nickname, msg.To, msg.Content)
		if err != nil {
			if _, ok := err.(service.ValidationError); !ok {
				log.Printf("ws: save message: %v", err)
			}
			return
		}
		response := Message{
			Type:      "new_message",
			ID:        out.ID,
			From:      out.From,
			FromNick:  out.FromNick,
			To:        out.To,
			Content:   out.Content,
			CreatedAt: out.CreatedAt,
		}
		h.SendToUser(msg.To, response)
		h.SendToUser(c.UserID, response)

	case "typing":
		if msg.To == 0 {
			return
		}
		h.SendToUser(msg.To, Message{
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

// BroadcastForum fans out forum events to every connected client (posts, comments, reactions).
func (h *Hub) BroadcastForum(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case h.broadcast <- data:
	default:
		log.Println("hub: broadcast queue full")
	}
}
