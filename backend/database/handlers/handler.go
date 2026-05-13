package handlers

import (
	"database/sql"
	"net/http"
	"path/filepath"

	"forum/backend/websocket"
)

type Handler struct {
	DB  *sql.DB
	Hub *websocket.Hub
}

func NewHandler(db *sql.DB, hub *websocket.Hub) *Handler {
	return &Handler{DB: db, Hub: hub}
}

// ServeIndex отдаёт единственный HTML-файл (SPA)
func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	// если запрос — не корень и не страница, не обслуживаем как HTML
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join("frontend", "index.html"))
}
