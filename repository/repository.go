package repository

import (
	"microblogging/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type PostRepository interface {
	Save(post *model.Post) (uuid.UUID, error)
	GetTimeline(userID string) ([]model.Post, error)
	FollowUser(followerID, followeeID string) error
	GetFollowees(userID string, limit int) ([]string, error)
	CreateUser(userData model.CreateUserRequest) (uuid.UUID, error)
	DeleteUser(userID string) error
	GetUser(userID string) (model.User, error)
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
func (p *postRepo) GetFollowees(userID string, limit int) ([]string, error) {
	panic("unimplemented")
}

// GetTimeline implements PostRepository.
func (p *postRepo) GetTimeline(userID string) ([]model.Post, error) {
	panic("unimplemented")
}

// Save implements PostRepository.
func (p *postRepo) Save(post *model.Post) (uuid.UUID, error) {
	panic("unimplemented")
}

// CreateUser implements PostRepository.
func (p *postRepo) CreateUser(userData model.CreateUserRequest) (uuid.UUID, error) {
	panic("unimplemented")
}

// DeleteUser implements PostRepository.
func (p *postRepo) DeleteUser(userID string) error {
	panic("unimplemented")
}

// GetUser implements PostRepository.
func (p *postRepo) GetUser(userID string) (model.User, error) {
	panic("unimplemented")
}

func NewPostRepository(db *sqlx.DB, logger *zap.Logger) PostRepository {
	return &postRepo{db: db, logger: logger}
}
