package repository

import (
	"database/sql"
	"errors"
	"time"

	"forum/internal/domain"
)

const (
	reactionLike    = 1
	reactionDislike = -1
)

var ErrPostNotFound = errors.New("post not found")

type ReactionRepository struct {
	db *sql.DB
}

func NewReactionRepository(db *sql.DB) *ReactionRepository {
	return &ReactionRepository{db: db}
}

// SetReaction applies like, dislike, or remove in a single transaction (toggle on repeat).
func (r *ReactionRepository) SetReaction(userID, postID int, action string) (domain.ReactionUpdate, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return domain.ReactionUpdate{}, err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRow(`SELECT 1 FROM posts WHERE id = ?`, postID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ReactionUpdate{}, ErrPostNotFound
		}
		return domain.ReactionUpdate{}, err
	}

	var current sql.NullInt64
	err = tx.QueryRow(
		`SELECT value FROM reactions WHERE user_id = ? AND post_id = ?`,
		userID, postID,
	).Scan(&current)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.ReactionUpdate{}, err
	}

	switch action {
	case "remove":
		_, err = tx.Exec(`DELETE FROM reactions WHERE user_id = ? AND post_id = ?`, userID, postID)
	case "like", "dislike":
		target := reactionLike
		if action == "dislike" {
			target = reactionDislike
		}
		if current.Valid && int(current.Int64) == target {
			_, err = tx.Exec(`DELETE FROM reactions WHERE user_id = ? AND post_id = ?`, userID, postID)
		} else if current.Valid {
			_, err = tx.Exec(
				`UPDATE reactions SET value = ?, updated_at = ? WHERE user_id = ? AND post_id = ?`,
				target, time.Now(), userID, postID,
			)
		} else {
			_, err = tx.Exec(
				`INSERT INTO reactions (user_id, post_id, value) VALUES (?, ?, ?)`,
				userID, postID, target,
			)
		}
	default:
		return domain.ReactionUpdate{}, errors.New("invalid reaction action")
	}
	if err != nil {
		return domain.ReactionUpdate{}, err
	}

	likes, dislikes, err := countsInTx(tx, postID)
	if err != nil {
		return domain.ReactionUpdate{}, err
	}

	userReaction, err := userReactionInTx(tx, userID, postID)
	if err != nil {
		return domain.ReactionUpdate{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.ReactionUpdate{}, err
	}

	return domain.ReactionUpdate{
		PostID:       postID,
		UserID:       userID,
		Reaction:     userReaction,
		LikeCount:    likes,
		DislikeCount: dislikes,
	}, nil
}

func (r *ReactionRepository) CountsForPost(postID int) (likes, dislikes int, err error) {
	return r.counts(r.db, postID)
}

func (r *ReactionRepository) UserReaction(userID, postID int) (string, error) {
	return userReactionInTx(r.db, userID, postID)
}

func countsInTx(q sqlQuerier, postID int) (likes, dislikes int, err error) {
	err = q.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END), 0)
		FROM reactions WHERE post_id = ?`, postID,
	).Scan(&likes, &dislikes)
	return
}

func userReactionInTx(q sqlQuerier, userID, postID int) (string, error) {
	var value sql.NullInt64
	err := q.QueryRow(
		`SELECT value FROM reactions WHERE user_id = ? AND post_id = ?`,
		userID, postID,
	).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if !value.Valid {
		return "", nil
	}
	switch int(value.Int64) {
	case reactionLike:
		return "like", nil
	case reactionDislike:
		return "dislike", nil
	default:
		return "", nil
	}
}

type sqlQuerier interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

func (r *ReactionRepository) counts(q sqlQuerier, postID int) (likes, dislikes int, err error) {
	return countsInTx(q, postID)
}
