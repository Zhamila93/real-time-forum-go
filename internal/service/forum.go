package service

import (
	"database/sql"
	"errors"
	"strings"

	"forum/internal/domain"
	"forum/internal/repository"
)

type ForumService struct {
	posts      *repository.PostRepository
	comments   *repository.CommentRepository
	categories *repository.CategoryRepository
	reactions  *repository.ReactionRepository
	realtime   ForumRealtime
}

func NewForumService(
	posts *repository.PostRepository,
	comments *repository.CommentRepository,
	categories *repository.CategoryRepository,
	reactions *repository.ReactionRepository,
	realtime ForumRealtime,
) *ForumService {
	return &ForumService{
		posts:      posts,
		comments:   comments,
		categories: categories,
		reactions:  reactions,
		realtime:   realtime,
	}
}

func (s *ForumService) ListCategories() ([]domain.Category, error) {
	return s.categories.List()
}

func (s *ForumService) ListPosts(category string) ([]domain.Post, error) {
	return s.posts.List(category)
}

type CreatePostInput struct {
	Title      string
	Content    string
	Categories []string
}

func (s *ForumService) CreatePost(userID int, in CreatePostInput) (domain.Post, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Content = strings.TrimSpace(in.Content)
	if in.Title == "" || in.Content == "" {
		return domain.Post{}, ValidationError{Message: "Заполните заголовок и текст"}
	}
	if len(in.Categories) == 0 {
		return domain.Post{}, ValidationError{Message: "Выберите хотя бы одну категорию"}
	}

	postID, err := s.posts.Create(userID, in.Title, in.Content, in.Categories)
	if err != nil {
		return domain.Post{}, err
	}

	post, err := s.posts.GetByID(int(postID))
	if err != nil {
		return domain.Post{}, err
	}

	if s.realtime != nil {
		s.realtime.PostCreated(post)
	}
	return post, nil
}

func (s *ForumService) GetPostWithComments(id, viewerID int) (domain.Post, []domain.Comment, error) {
	post, err := s.posts.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Post{}, nil, ErrNotFound
		}
		return domain.Post{}, nil, err
	}
	comments, err := s.comments.ListByPostID(id)
	if err != nil {
		return domain.Post{}, nil, err
	}
	if viewerID > 0 {
		post.UserReaction, _ = s.reactions.UserReaction(viewerID, id)
	}
	return post, comments, nil
}

func (s *ForumService) CreateComment(userID, postID int, content string) (domain.Comment, error) {
	content = strings.TrimSpace(content)
	if content == "" || postID == 0 {
		return domain.Comment{}, ValidationError{Message: "Введите комментарий"}
	}

	if _, err := s.posts.GetByID(postID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Comment{}, ErrNotFound
		}
		return domain.Comment{}, err
	}

	commentID, err := s.comments.Create(postID, userID, content)
	if err != nil {
		return domain.Comment{}, err
	}

	comment, err := s.comments.GetByID(int(commentID))
	if err != nil {
		return domain.Comment{}, err
	}

	if s.realtime != nil {
		s.realtime.CommentCreated(comment)
	}
	return comment, nil
}

// SetReaction toggles like/dislike or removes the user's reaction atomically.
func (s *ForumService) SetReaction(userID, postID int, action string) (domain.ReactionUpdate, error) {
	action = strings.ToLower(strings.TrimSpace(action))
	switch action {
	case "like", "dislike", "remove":
	default:
		return domain.ReactionUpdate{}, ValidationError{Message: "reaction must be like, dislike, or remove"}
	}
	if postID <= 0 {
		return domain.ReactionUpdate{}, ValidationError{Message: "invalid post id"}
	}

	update, err := s.reactions.SetReaction(userID, postID, action)
	if errors.Is(err, repository.ErrPostNotFound) {
		return domain.ReactionUpdate{}, ErrNotFound
	}
	return update, err
}
