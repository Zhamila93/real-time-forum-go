package websocket

import (
	"forum/internal/domain"
	"forum/internal/service"
)

// ForumBroadcaster adapts Hub to service.ForumRealtime.
type ForumBroadcaster struct {
	hub *Hub
}

func NewForumBroadcaster(hub *Hub) *ForumBroadcaster {
	return &ForumBroadcaster{hub: hub}
}

var _ service.ForumRealtime = (*ForumBroadcaster)(nil)

func (b *ForumBroadcaster) PostCreated(post domain.Post) {
	b.hub.BroadcastForum(PostCreatedMessage(post))
}

func (b *ForumBroadcaster) CommentCreated(comment domain.Comment) {
	b.hub.BroadcastForum(CommentCreatedMessage(comment))
}
