package repository

import (
	"fmt"
	"microblogging/model"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type DBConnector struct {
	DB     *sqlx.DB
	Logger *zap.Logger
}

func (r *DBConnector) NewPostRepository() PostRepository {
	return &postRepo{db: r.DB, logger: r.Logger}
}

func (r *DBConnector) Save(post *model.Post) (uuid.UUID, error) {
	var postID uuid.UUID

	insertQuery := `
		INSERT INTO posts (user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, content)
		DO UPDATE SET 
			content = EXCLUDED.content,
			updated_at = EXCLUDED.updated_at
		RETURNING id;
	`

	err := r.DB.QueryRow(insertQuery,
		post.UserID, post.Content, post.CreatedAt, post.UpdatedAt,
	).Scan(&postID)
	if err != nil {
		r.Logger.Error("Error saving/updating post", zap.Error(err))
		return uuid.Nil, err
	}

	// Goroutine is not blocker if user update is failed
	go func(postID uuid.UUID, userID string, updatedAt time.Time) {
		updateUserQuery := `
			UPDATE users
			SET last_post_id = $1, updated_at = $2
			WHERE id = $3;
		`
		if _, err := r.DB.Exec(updateUserQuery, postID, updatedAt, userID); err != nil {
			r.Logger.Error("Error updating user's last_post_id", zap.Error(err))
		}
	}(postID, post.UserID, post.UpdatedAt)

	r.Logger.Sugar().Infow("Post saved", "post_id", postID.String())
	return postID, nil
}

func (r *DBConnector) GetTimeline(userID string) ([]model.Post, error) {
	var posts []model.Post
	query := `
		SELECT p.id, p.user_id, p.content, p.created_at
		FROM posts p
		JOIN follows f ON f.followee_id = p.user_id
		WHERE f.follower_id = $1
		ORDER BY p.created_at DESC
	`
	err := r.DB.Select(&posts, query, userID)
	if err != nil {
		r.Logger.Error("Error getting timeline", zap.Error(err))
		return nil, err
	}
	return posts, nil
}

func (r *DBConnector) FollowUser(followerID, followeeID string) error {
	query := `
		INSERT INTO follows (follower_id, followee_id) 
		VALUES ($1, $2) 
		ON CONFLICT DO NOTHING
	`
	_, err := r.DB.Exec(query, followerID, followeeID)
	if err != nil {
		r.Logger.Error("Error following user", zap.Error(err))
	}
	return err
}

func (r *DBConnector) GetFollowees(userID string) ([]string, error) {
	var followees []string
	query := `SELECT followee_id FROM follows WHERE follower_id = $1`
	err := r.DB.Select(&followees, query, userID)
	if err != nil {
		r.Logger.Error("Error getting followees", zap.Error(err))
		return nil, err
	}
	r.Logger.Sugar().Infow("Got followees info", "user_id", userID)
	return followees, nil
}

func (r *DBConnector) CreateUser(userData model.CreateUserRequest) (uuid.UUID, error) {
	now := time.Now().UTC()
	var userID uuid.UUID
	query := `
		INSERT INTO users (username, created_at, updated_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err := r.DB.QueryRow(query, userData.Name, now, now).Scan(&userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to insert user: %w", err)
	}
	r.Logger.Sugar().Infow("User created successfully", "user_id", userID.String())
	return userID, nil
}

func (r *DBConnector) GetUser(userID string) (model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE id = $1`
	if err := r.DB.Get(&user, query, userID); err != nil {
		r.Logger.Error("Error getting user", zap.Error(err))
		return model.User{}, err
	}
	r.Logger.Sugar().Infow("Got user info", "user_id", userID)
	return user, nil
}

func (r *DBConnector) DeleteUser(userID string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.DB.Exec(query, userID)
	if err != nil {
		r.Logger.Error("Error deleting user", zap.Error(err))
	}
	r.Logger.Sugar().Info("User is deleted", "user_id", userID)
	return err
}
