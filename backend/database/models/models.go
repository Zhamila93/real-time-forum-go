package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Nickname  string    `json:"nickname"`
	Age       int       `json:"age"`
	Gender    string    `json:"gender"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Post struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Nickname     string    `json:"nickname"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Categories   []string  `json:"categories"`
	CreatedAt    time.Time `json:"created_at"`
	CommentCount int       `json:"comment_count"`
}

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	Sender     string    `json:"sender"`
	Receiver   string    `json:"receiver"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserListItem struct {
	ID          int        `json:"id"`
	Nickname    string     `json:"nickname"`
	Online      bool       `json:"online"`
	LastMessage *time.Time `json:"last_message"`
}
