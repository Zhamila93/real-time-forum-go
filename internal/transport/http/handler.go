package http

import (
	"net/http"
	"path/filepath"

	"forum/internal/config"
	"forum/internal/service"
	ws "forum/internal/transport/websocket"
)

type Handler struct {
	Auth   *service.AuthService
	Forum  *service.ForumService
	Chat   *service.ChatService
	Hub    *ws.Hub
	Config config.Config
}

func NewHandler(
	auth *service.AuthService,
	forum *service.ForumService,
	chat *service.ChatService,
	hub *ws.Hub,
	cfg config.Config,
) *Handler {
	return &Handler{Auth: auth, Forum: forum, Chat: chat, Hub: hub, Config: cfg}
}

func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(h.Config.StaticDir, "index.html"))
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	fs := http.FileServer(http.Dir(h.Config.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", h.ServeIndex)

	mux.HandleFunc("/api/register", h.Register)
	mux.HandleFunc("/api/login", h.Login)
	mux.HandleFunc("/api/logout", h.Logout)
	mux.HandleFunc("/api/me", h.Me)

	mux.HandleFunc("/api/posts", h.Posts)
	mux.HandleFunc("/api/posts/", h.PostSubroutes)
	mux.HandleFunc("/api/comments", h.CreateComment)
	mux.HandleFunc("/api/categories", h.Categories)

	mux.HandleFunc("/api/users", h.UsersList)
	mux.HandleFunc("/api/messages", h.GetMessages)

	mux.HandleFunc("/ws", h.WSHandler)
}
