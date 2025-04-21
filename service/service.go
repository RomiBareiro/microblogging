package service

import (
	m "microblogging/model"
	"microblogging/repository"
	"time"

	"github.com/google/uuid"
)

type BlogService interface {
	CreatePost(userID, content string) (uuid.UUID, error)
	GetTimeline(timeLine m.TimelineRequest) (m.TimelineResponse, error)
	FollowUser(followerID, followeeID string) error
	UnfollowUser(followerID, followeeID string) error
	GetFollowees(userID string, limit int) ([]string, error)
	CreateUser(userData m.CreateUserRequest) (uuid.UUID, error)
	UpdatePostPut(post m.CreatePostRequest) error
	DeleteUser(userID string) error
	GetUser(userID string) (m.User, error)
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

func (s *blogService) GetTimeline(info m.TimelineRequest) (m.TimelineResponse, error) {
	return s.repo.GetTimeline(info)
}

func (s *blogService) FollowUser(followerID, followeeID string) error {
	return s.repo.FollowUser(followerID, followeeID)
}
func (s *blogService) UnfollowUser(followerID, followeeID string) error {
	return s.repo.UnfollowUser(followerID, followeeID)
}

func (s *blogService) GetFollowees(userID string, limit int) ([]string, error) {
	return s.repo.GetFollowees(userID, limit)
}

func (s *blogService) CreateUser(userData m.CreateUserRequest) (uuid.UUID, error) {
	return s.repo.CreateUser(userData)
}
func (s *blogService) UpdatePostPut(post m.CreatePostRequest) error {
	return s.repo.UpdatePostPut(post)
}

func (s *blogService) DeleteUser(userID string) error {
	return s.repo.DeleteUser(userID)
}

func (s *blogService) GetUser(userID string) (m.User, error) {
	return s.repo.GetUser(userID)
}
