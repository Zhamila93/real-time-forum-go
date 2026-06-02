package service

import "forum/internal/domain"

// ForumRealtime pushes forum events to connected WebSocket clients.
type ForumRealtime interface {
	PostCreated(post domain.Post)
	CommentCreated(comment domain.Comment)
}
