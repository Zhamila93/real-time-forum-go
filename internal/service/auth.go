package service

import (
	"database/sql"
	"strings"
	"time"

	"forum/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const SessionDuration = 24 * time.Hour * 7

type SessionInfo struct {
	ID        string
	UserID    int
	ExpiresAt time.Time
}

type AuthService struct {
	users    *repository.UserRepository
	sessions *repository.SessionRepository
}

func NewAuthService(users *repository.UserRepository, sessions *repository.SessionRepository) *AuthService {
	return &AuthService{users: users, sessions: sessions}
}

type RegisterInput struct {
	Nickname  string
	Age       int
	Gender    string
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type AuthResult struct {
	UserID   int
	Nickname string
	Session  SessionInfo
}

func (s *AuthService) Register(in RegisterInput) (AuthResult, error) {
	in.Nickname = strings.TrimSpace(in.Nickname)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)
	in.Gender = strings.TrimSpace(in.Gender)

	if in.Nickname == "" || in.Email == "" || in.Password == "" ||
		in.FirstName == "" || in.LastName == "" || in.Gender == "" {
		return AuthResult{}, ValidationError{Message: "Все поля обязательны"}
	}
	if in.Age < 1 || in.Age > 150 {
		return AuthResult{}, ValidationError{Message: "Некорректный возраст"}
	}
	if len(in.Password) < 6 {
		return AuthResult{}, ValidationError{Message: "Пароль минимум 6 символов"}
	}
	if !strings.Contains(in.Email, "@") {
		return AuthResult{}, ValidationError{Message: "Некорректный e-mail"}
	}

	exists, err := s.users.ExistsByNicknameOrEmail(in.Nickname, in.Email)
	if err != nil {
		return AuthResult{}, err
	}
	if exists {
		return AuthResult{}, ValidationError{Message: "Никнейм или e-mail уже заняты"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, err
	}

	userID, err := s.users.Create(in.Nickname, in.Age, in.Gender, in.FirstName, in.LastName, in.Email, string(hash))
	if err != nil {
		return AuthResult{}, err
	}

	sess, err := s.createSession(int(userID))
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		UserID:   int(userID),
		Nickname: in.Nickname,
		Session:  sess,
	}, nil
}

type LoginInput struct {
	Identifier string
	Password   string
}

func (s *AuthService) Login(in LoginInput) (AuthResult, error) {
	in.Identifier = strings.TrimSpace(in.Identifier)
	if in.Identifier == "" || in.Password == "" {
		return AuthResult{}, ValidationError{Message: "Введите логин и пароль"}
	}

	id, nickname, hash, err := s.users.FindByIdentifier(in.Identifier, strings.ToLower(in.Identifier))
	if err != nil {
		if err == sql.ErrNoRows {
			return AuthResult{}, ErrUnauthorized
		}
		return AuthResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)); err != nil {
		return AuthResult{}, ErrUnauthorized
	}

	sess, err := s.createSession(id)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{UserID: id, Nickname: nickname, Session: sess}, nil
}

func (s *AuthService) Logout(sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return s.sessions.Delete(sessionID)
}

func (s *AuthService) UserIDFromSession(sessionID string) int {
	if sessionID == "" {
		return 0
	}
	userID, expiresAt, err := s.sessions.Lookup(sessionID)
	if err != nil {
		return 0
	}
	if time.Now().After(expiresAt) {
		_ = s.sessions.Delete(sessionID)
		return 0
	}
	return userID
}

func (s *AuthService) Me(userID int) (int, string, error) {
	nickname, err := s.users.NicknameByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", ErrNotFound
		}
		return 0, "", err
	}
	return userID, nickname, nil
}

func (s *AuthService) createSession(userID int) (SessionInfo, error) {
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(SessionDuration)

	_ = s.sessions.DeleteByUserID(userID)

	if err := s.sessions.Create(sessionID, userID, expiresAt); err != nil {
		return SessionInfo{}, err
	}

	return SessionInfo{ID: sessionID, UserID: userID, ExpiresAt: expiresAt}, nil
}
