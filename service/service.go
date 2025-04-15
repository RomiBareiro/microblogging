package service

import (
	"microblogging/model"
	"microblogging/repository"
	"time"

	"github.com/google/uuid"
)

type BlogService interface {
	CreatePost(userID, content string) error
	GetTimeline(userID string) ([]model.Post, error)
	FollowUser(followerID, followeeID string) error
	GetFollowees(userID string) ([]string, error)
}

type blogService struct {
	repo repository.PostRepository
}

func NewBlogService(r repository.PostRepository) BlogService {
	return &blogService{repo: r}
}

func (s *blogService) CreatePost(userID, content string) error {
	post := &model.Post{
		ID:        uuid.New().String(),
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}
	return s.repo.Save(post)
}

func (s *blogService) GetTimeline(userID string) ([]model.Post, error) {
	return s.repo.GetTimeline(userID)
}

func (s *blogService) FollowUser(followerID, followeeID string) error {
	return s.repo.FollowUser(followerID, followeeID)
}

func (s *blogService) GetFollowees(userID string) ([]string, error) {
	return s.repo.GetFollowees(userID)
}
