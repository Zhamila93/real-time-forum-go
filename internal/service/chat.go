package service

import (
	"sort"
	"time"

	"forum/internal/domain"
	"forum/internal/repository"
)

type OnlineProvider interface {
	OnlineUsersMap() map[int]bool
}

type ChatService struct {
	users    *repository.UserRepository
	messages *repository.MessageRepository
}

func NewChatService(users *repository.UserRepository, messages *repository.MessageRepository) *ChatService {
	return &ChatService{users: users, messages: messages}
}

func (s *ChatService) ListUsers(currentUserID int, online OnlineProvider) ([]domain.UserListItem, error) {
	rows, err := s.users.ListExcept(currentUserID)
	if err != nil {
		return nil, err
	}

	onlineMap := map[int]bool{}
	if online != nil {
		onlineMap = online.OnlineUsersMap()
	}

	var withMsg, withoutMsg []domain.UserListItem
	for _, row := range rows {
		item := domain.UserListItem{
			ID:          row.ID,
			Nickname:    row.Nickname,
			Online:      onlineMap[row.ID],
			LastMessage: row.LastMessage,
		}
		if row.LastMessage != nil {
			withMsg = append(withMsg, item)
		} else {
			withoutMsg = append(withoutMsg, item)
		}
	}

	sort.Slice(withMsg, func(i, j int) bool {
		if withMsg[i].LastMessage == nil {
			return false
		}
		if withMsg[j].LastMessage == nil {
			return true
		}
		return withMsg[i].LastMessage.After(*withMsg[j].LastMessage)
	})
	sort.Slice(withoutMsg, func(i, j int) bool {
		return withoutMsg[i].Nickname < withoutMsg[j].Nickname
	})

	result := make([]domain.UserListItem, 0, len(withMsg)+len(withoutMsg))
	result = append(result, withMsg...)
	result = append(result, withoutMsg...)
	return result, nil
}

func (s *ChatService) GetMessages(currentUserID, withID, limit, offset int) ([]domain.Message, error) {
	if withID == 0 {
		return nil, ValidationError{Message: "Некорректный параметр with"}
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.messages.ListBetween(currentUserID, withID, limit, offset)
}

func (s *ChatService) SendPrivateMessage(senderID int, senderNick string, to int, content string) (domain.ChatOutbound, error) {
	if to == 0 || content == "" {
		return domain.ChatOutbound{}, ValidationError{Message: "invalid message"}
	}
	id, err := s.messages.Create(senderID, to, content)
	if err != nil {
		return domain.ChatOutbound{}, err
	}
	return domain.ChatOutbound{
		ID:        int(id),
		From:      senderID,
		FromNick:  senderNick,
		To:        to,
		Content:   content,
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}
