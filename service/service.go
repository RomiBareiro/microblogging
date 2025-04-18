package service

import (
	m "microblogging/model"
	"microblogging/repository"
	"time"

	"github.com/google/uuid"
)

type BlogService interface {
	CreatePost(userID, content string) (uuid.UUID, error)
	GetTimeline(userID string) ([]m.Post, error)
	FollowUser(followerID, followeeID string) error
	GetFollowees(userID string, limit int) ([]string, error)
	CreateUser(userData m.CreateUserRequest) (uuid.UUID, error)
}

type blogService struct {
	repo repository.PostRepository
}

func NewBlogService(r repository.PostRepository) BlogService {
	return &blogService{repo: r}
}

func (s *blogService) CreatePost(userID, content string) (uuid.UUID, error) {
	post := &m.Post{
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}
	return s.repo.Save(post)
}

func (s *blogService) GetTimeline(userID string) ([]m.Post, error) {
	return s.repo.GetTimeline(userID)
}

func (s *blogService) FollowUser(followerID, followeeID string) error {
	return s.repo.FollowUser(followerID, followeeID)
}

func (s *blogService) GetFollowees(userID string, limit int) ([]string, error) {
	return s.repo.GetFollowees(userID, limit)
}

func (s *blogService) CreateUser(userData m.CreateUserRequest) (uuid.UUID, error) {
	return s.repo.CreateUser(userData)
}
