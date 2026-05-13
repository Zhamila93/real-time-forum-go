package main

import (
	"log"
	"net/http"

	"forum/backend/database"
	"forum/backend/handlers"
	"forum/backend/websocket"
)

func main() {
	// Инициализация БД
	db, err := database.InitDB("./forum.db")
	if err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}
	defer db.Close()

	// Создаём хаб для WebSocket
	hub := websocket.NewHub(db)
	go hub.Run()

	// Создаём обработчики
	h := handlers.NewHandler(db, hub)

	// Статика — фронтенд
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Главная страница (SPA)
	http.HandleFunc("/", h.ServeIndex)

	// API-эндпоинты
	http.HandleFunc("/api/register", h.Register)
	http.HandleFunc("/api/login", h.Login)
	http.HandleFunc("/api/logout", h.Logout)
	http.HandleFunc("/api/me", h.Me)

	http.HandleFunc("/api/posts", h.Posts)            // GET список, POST создать
	http.HandleFunc("/api/posts/", h.PostByID)        // GET один пост с комментариями
	http.HandleFunc("/api/comments", h.CreateComment) // POST создать коммент
	http.HandleFunc("/api/categories", h.Categories)

	http.HandleFunc("/api/users", h.UsersList)
	http.HandleFunc("/api/messages", h.GetMessages)

	// WebSocket
	http.HandleFunc("/ws", h.WSHandler)

	log.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
