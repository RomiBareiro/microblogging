package repository

import (
	"microblogging/model"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type PostRepository interface {
	Save(post *model.Post) error
	GetTimeline(userID string) ([]model.Post, error)
	FollowUser(followerID, followeeID string) error
	GetFollowees(userID string) ([]string, error)
}

type postRepo struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// FollowUser implements PostRepository.
func (p *postRepo) FollowUser(followerID string, followeeID string) error {
	panic("unimplemented")
}

// GetFollowees implements PostRepository.
func (p *postRepo) GetFollowees(userID string) ([]string, error) {
	panic("unimplemented")
}

// GetTimeline implements PostRepository.
func (p *postRepo) GetTimeline(userID string) ([]model.Post, error) {
	panic("unimplemented")
}

// Save implements PostRepository.
func (p *postRepo) Save(post *model.Post) error {
	panic("unimplemented")
}

func NewPostRepository(db *sqlx.DB, logger *zap.Logger) PostRepository {
	return &postRepo{db: db, logger: logger}
}
