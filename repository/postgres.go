package repository

import (
	"microblogging/model"

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

func (r *DBConnector) Save(post *model.Post) error {
	_, err := r.DB.Exec(`INSERT INTO posts (id, user_id, content, created_at)
		VALUES ($1, $2, $3, $4)`,
		post.ID, post.UserID, post.Content, post.CreatedAt)
	if err != nil {
		r.Logger.Error("Error saving post", zap.Error(err))
	}
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

func (r *DBConnector) GetFollowees(userID string) ([]string, error) {
	var followees []string
	query := `SELECT followee_id FROM follows WHERE follower_id = $1`
	err := r.DB.Select(&followees, query, userID)
	if err != nil {
		r.Logger.Error("Error getting followees", zap.Error(err))
		return nil, err
	}
	return followees, nil
}
