package main

import (
	"log"
	"net/http"

	"forum/internal/config"
	"forum/internal/database"
	"forum/internal/repository"
	"forum/internal/service"
	httpx "forum/internal/transport/http"
	"forum/internal/transport/websocket"
)

func main() {
	cfg := config.Default()

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatal("database:", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	reactionRepo := repository.NewReactionRepository(db)

	authSvc := service.NewAuthService(userRepo, sessionRepo)
	chatSvc := service.NewChatService(userRepo, messageRepo)

	hub := websocket.NewHub(chatSvc)
	go hub.Run()

	forumRealtime := websocket.NewForumBroadcaster(hub)
	forumSvc := service.NewForumService(postRepo, commentRepo, categoryRepo, reactionRepo, forumRealtime)

	handler := httpx.NewHandler(authSvc, forumSvc, chatSvc, hub, cfg)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	log.Println("Server listening on http://localhost" + cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		log.Fatal(err)
	}
}
