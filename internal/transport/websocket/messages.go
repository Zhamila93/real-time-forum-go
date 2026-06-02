package websocket

import "forum/internal/domain"

// Message is the WebSocket JSON envelope shared with the browser client.
type Message struct {
	Type      string `json:"type"`
	ID        int    `json:"id,omitempty"`
	From      int    `json:"from,omitempty"`
	FromNick  string `json:"from_nick,omitempty"`
	To        int    `json:"to,omitempty"`
	Content   string `json:"content,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	OnlineIDs []int  `json:"online_ids,omitempty"`

	// Forum events (posts, comments, reactions)
	PostID       int             `json:"post_id,omitempty"`
	UserID       int             `json:"user_id,omitempty"`
	Reaction     string          `json:"reaction,omitempty"`
	LikeCount    int             `json:"like_count,omitempty"`
	DislikeCount int             `json:"dislike_count,omitempty"`
	Post         *domain.Post    `json:"post,omitempty"`
	Comment      *domain.Comment `json:"comment,omitempty"`
}

func PostCreatedMessage(p domain.Post) Message {
	post := p
	return Message{Type: "post_created", Post: &post}
}

func CommentCreatedMessage(c domain.Comment) Message {
	comment := c
	return Message{Type: "comment_created", PostID: c.PostID, Comment: &comment}
}

func ReactionUpdatedMessage(u domain.ReactionUpdate) Message {
	return Message{
		Type:         "reaction_updated",
		PostID:       u.PostID,
		UserID:       u.UserID,
		Reaction:     u.Reaction,
		LikeCount:    u.LikeCount,
		DislikeCount: u.DislikeCount,
	}
}
