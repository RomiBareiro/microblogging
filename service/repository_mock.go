package service

import (
	"fmt"
	"microblogging/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Save(post *model.Post) (uuid.UUID, error) {
	args := m.Called(post)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockPostRepository) GetTimeline(info model.TimelineRequest) (model.TimelineResponse, error) {
	args := m.Called(info)
	return args.Get(0).(model.TimelineResponse), args.Error(1)
}

func (m *MockPostRepository) FollowUser(followerID, followeeID string) error {
	args := m.Called(followerID, followeeID)
	return args.Error(0)
}

func (m *MockPostRepository) UnfollowUser(followerID, followeeID string) error {
	args := m.Called(followerID, followeeID)
	return args.Error(0)
}
func (m *MockPostRepository) GetFollowees(userID string, limit int) ([]string, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPostRepository) CreateUser(userData model.CreateUserRequest) (uuid.UUID, error) {
	args := m.Called(userData)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockPostRepository) UpdatePostPut(post model.CreatePostRequest) error {
	fmt.Println("Mock called with:", post)

	args := m.Called(post)
	return args.Error(0)
}

func (m *MockPostRepository) DeleteUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockPostRepository) GetUser(userID string) (model.User, error) {
	args := m.Called(userID)
	return args.Get(0).(model.User), args.Error(1)
}
