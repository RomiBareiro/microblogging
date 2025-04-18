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
	now := time.Now().UTC()

	const insertQuery = `
		INSERT INTO posts (user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	err := r.DB.QueryRow(insertQuery, post.UserID, post.Content, now, now).Scan(&postID)
	if err != nil {
		r.Logger.Error("Error inserting post", zap.Error(err))
		return uuid.Nil, err
	}
	// Update the user's last post asynchronously
	r.updateUserLastPostAsync(postID, post.UserID, now)

	r.Logger.Sugar().Infow("Post saved", "post_id", postID.String())
	return postID, nil
}

func (r *DBConnector) UpdatePostPut(post model.CreatePostRequest) error {
	now := time.Now().UTC()
	postUUID, err := uuid.Parse(post.PostID)
	if err != nil {
		r.Logger.Error("Invalid post_id UUID", zap.Error(err))
		return model.ErrInvalidUUID
	}

	if err := r.existPost(postUUID, post.UserID); err != nil {
		return model.ErrPostNotFound
	}

	const updateQuery = `
		UPDATE posts
		SET content = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4;
	`
	_, err = r.DB.Exec(updateQuery, post.Content, now, post.PostID, post.UserID)
	if err != nil {
		r.Logger.Error("Error updating post", zap.Error(err))
		return err
	}

	// Update the user's last post asynchronously
	r.updateUserLastPostAsync(postUUID, post.UserID, now)
	r.Logger.Sugar().Infow("Post updated", "post_id", post.PostID)
	return err
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

func (r *DBConnector) GetFollowees(userID string, limit int) ([]string, error) {
	var followees []string
	query := `SELECT followee_id
			  FROM follows 
			  WHERE follower_id = $1
			  LIMIT $2`
	err := r.DB.Select(&followees, query, userID, limit)
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
		INSERT INTO users (user_name, created_at, updated_at)
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
		return err
	}
	r.Logger.Sugar().Info("User is deleted", "user_id", userID)
	return err
}

func (r *DBConnector) updateUserLastPostAsync(postID uuid.UUID, userID string, updatedAt time.Time) {
	go func() {
		const updateUserQuery = `
			UPDATE users
			SET last_post_id = $1, updated_at = $2
			WHERE id = $3;
		`
		if _, err := r.DB.Exec(updateUserQuery, postID, updatedAt, userID); err != nil {
			r.Logger.Error("Error updating user's last_post_id", zap.Error(err))
		}
		r.Logger.Sugar().Infow("User's last_post_id updated", "user_id", userID, "post_id", postID.String())
	}()
}

func (r *DBConnector) existPost(postID uuid.UUID, userID string) error {
	var exists bool
	checkPostQuery := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1 AND user_id = $2);`
	err := r.DB.QueryRow(checkPostQuery, postID, userID).Scan(&exists)
	if err != nil {
		r.Logger.Error("Error checking if post exists", zap.Error(err))
		return err
	}

	if !exists {
		r.Logger.Sugar().Errorw("post_id does not exist for user_id", "post_id", postID.String(), "user_id", userID)
		return model.ErrPostNotFound
	}
	r.Logger.Sugar().Infow("Existing post", "post_id", postID.String(), "user_id", userID)
	return nil
}
