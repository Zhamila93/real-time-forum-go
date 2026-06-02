package domain

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
	LikeCount    int       `json:"like_count"`
	DislikeCount int       `json:"dislike_count"`
	UserReaction string    `json:"user_reaction,omitempty"` // "", "like", "dislike"
}

// ReactionUpdate is broadcast after a like/dislike/remove.
type ReactionUpdate struct {
	PostID       int    `json:"post_id"`
	UserID       int    `json:"user_id"`
	Reaction     string `json:"reaction"`
	LikeCount    int    `json:"like_count"`
	DislikeCount int    `json:"dislike_count"`
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

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ChatOutbound is a persisted private message ready for WebSocket delivery.
type ChatOutbound struct {
	ID        int
	From      int
	FromNick  string
	To        int
	Content   string
	CreatedAt string
}
