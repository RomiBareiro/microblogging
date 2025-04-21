package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"microblogging/model"
	"microblogging/server"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService implements BlogService using testify/mock
type MockService struct {
	mock.Mock
}

// CreatePost mocks CreatePost method
func (m *MockService) CreatePost(userID string, content string) (uuid.UUID, error) {
	args := m.Called(userID, content)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// CreateUser mocks CreateUser method
func (m *MockService) CreateUser(userData model.CreateUserRequest) (uuid.UUID, error) {
	args := m.Called(userData)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// DeleteUser mocks DeleteUser method
func (m *MockService) DeleteUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// FollowUser mocks FollowUser method
func (m *MockService) FollowUser(followerID string, followeeID string) error {
	args := m.Called(followeeID, followerID)
	return args.Error(0)
}

// UnfollowUser mocks FollowUser method
func (m *MockService) UnfollowUser(followerID string, followeeID string) error {
	args := m.Called(followeeID, followerID)
	return args.Error(0)
}

// GetFollowees mocks GetFollowees method
func (m *MockService) GetFollowees(userID string, limit int) ([]string, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]string), args.Error(1)
}

// GetTimeline mocks GetTimeline method
func (m *MockService) GetTimeline(req model.TimelineRequest) (model.TimelineResponse, error) {
	args := m.Called(req)
	return args.Get(0).(model.TimelineResponse), args.Error(1)
}

// GetUser mocks GetUser method
func (m *MockService) GetUser(userID string) (model.User, error) {
	args := m.Called(userID)
	return args.Get(0).(model.User), args.Error(1)
}

// UpdatePostPut mocks UpdatePostPut method
func (m *MockService) UpdatePostPut(post model.CreatePostRequest) error {
	args := m.Called(post)
	return args.Error(0)
}

func TestCreatePostHandler(t *testing.T) {
	mockSvc := new(MockService)
	s := server.NewServer(context.Background(), mockSvc)
	validUserID := uuid.New().String()
	validContent := "Hello world"

	tests := []struct {
		name           string
		method         string
		body           interface{}
		mockReturnID   uuid.UUID
		mockReturnErr  error
		expectedStatus int
	}{
		{
			name:           "Method Not Allowed",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid JSON Body",
			method:         http.MethodPost,
			body:           "invalid-json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Content Too Long",
			method:         http.MethodPost,
			body:           model.CreatePostRequest{UserID: validUserID, Content: string(make([]byte, 1001))},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Service Error",
			method:         http.MethodPost,
			body:           model.CreatePostRequest{UserID: validUserID, Content: validContent},
			mockReturnID:   uuid.Nil,
			mockReturnErr:  errors.New("mock error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Success",
			method:         http.MethodPost,
			body:           model.CreatePostRequest{UserID: validUserID, Content: validContent},
			mockReturnID:   uuid.New(),
			mockReturnErr:  nil,
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc.ExpectedCalls = nil // clear previous expectations

			var body []byte
			if tt.body != nil {
				switch b := tt.body.(type) {
				case string:
					body = []byte(b)
				default:
					body, _ = json.Marshal(tt.body)
				}
			}

			// Setup mock expectation only if input is valid and method is POST
			if req, ok := tt.body.(model.CreatePostRequest); ok &&
				tt.method == http.MethodPost &&
				len(req.Content) <= 1000 &&
				tt.expectedStatus != http.StatusBadRequest { // case "Content Too Long"
				mockSvc.On("CreatePost", req.UserID, req.Content).Return(tt.mockReturnID, tt.mockReturnErr)
			}

			req := httptest.NewRequest(tt.method, "/posts", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			s.CreatePostHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestUpdatePostPutHandler(t *testing.T) {
	mockSvc := new(MockService)
	s := server.NewServer(context.Background(), mockSvc)

	validUserID := uuid.New().String()
	validPostID := uuid.New().String()
	validContent := "Updated content"

	tests := []struct {
		name           string
		method         string
		body           interface{}
		mockReturnErr  error
		expectedStatus int
	}{
		{
			name:           "Method Not Allowed",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid JSON Body",
			method:         http.MethodPut,
			body:           "not-json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing PostID",
			method:         http.MethodPut,
			body:           model.CreatePostRequest{UserID: validUserID, Content: validContent},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Service Error",
			method:         http.MethodPut,
			body:           model.CreatePostRequest{PostID: validPostID, UserID: validUserID, Content: validContent},
			mockReturnErr:  errors.New("mock update error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Success",
			method:         http.MethodPut,
			body:           model.CreatePostRequest{PostID: validPostID, UserID: validUserID, Content: validContent},
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc.ExpectedCalls = nil // clear previous expectations

			body, _ := json.Marshal(tt.body)

			// Setup mock expectation only for valid payloads
			if req, ok := tt.body.(model.CreatePostRequest); ok && req.PostID != "" && tt.method == http.MethodPut {
				mockSvc.On("UpdatePostPut", req).Return(tt.mockReturnErr)
			}

			req := httptest.NewRequest(tt.method, "/posts", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			s.UpdatePostPutHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}
func TestUnfollowHandler(t *testing.T) {
	mockSvc := new(MockService)
	s := server.NewServer(context.Background(), mockSvc)

	// fixed ids
	const validFollowerID string = "550e8400-e29b-41d4-a716-446655440000"
	const validFolloweeID string = "550e8400-e29b-41d4-a716-446655440001"

	tests := []struct {
		name           string
		method         string
		body           interface{}
		mockReturnErr  error
		expectedStatus int
	}{
		{
			name:           "Method Not Allowed",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid JSON Body",
			method:         http.MethodPost,
			body:           "not-json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing Required Fields",
			method:         http.MethodPost,
			body:           map[string]string{"follower_id": validFollowerID},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Service Error",
			method: http.MethodPost,
			body: map[string]string{
				"follower_id": validFollowerID,
				"followee_id": validFolloweeID,
			},
			mockReturnErr:  errors.New("mock unfollow error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "Success",
			method: http.MethodPost,
			body: map[string]string{
				"follower_id": validFollowerID,
				"followee_id": validFolloweeID,
			},
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Unfollow_Self",
			method: http.MethodPost,
			body: map[string]string{
				"follower_id": validFollowerID,
				"followee_id": validFollowerID,
			},
			mockReturnErr:  model.ErrCanNotFollowSelf,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc.ExpectedCalls = nil // reset expectations

			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			if m, ok := tt.body.(map[string]string); ok &&
				m["follower_id"] == validFollowerID &&
				m["followee_id"] == validFolloweeID &&
				tt.method == http.MethodPost {
				mockSvc.On("UnfollowUser", validFolloweeID, validFollowerID).Return(tt.mockReturnErr)
			}

			req := httptest.NewRequest(tt.method, "/unfollow", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			s.UnfollowUserHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}
func TestFollowHandler(t *testing.T) {
	mockSvc := new(MockService)
	s := server.NewServer(context.Background(), mockSvc)

	const validFollowerID = "550e8400-e29b-41d4-a716-446655440000"
	const validFolloweeID = "550e8400-e29b-41d4-a716-446655440001"

	tests := []struct {
		name           string
		method         string
		body           interface{}
		mockReturnErr  error
		expectedStatus int
	}{
		{
			name:           "Method Not Allowed",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid JSON Body",
			method:         http.MethodPost,
			body:           "invalid-json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing Required Fields",
			method:         http.MethodPost,
			body:           map[string]string{"follower_id": validFollowerID},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Follow Self",
			method: http.MethodPost,
			body: map[string]string{
				"follower_id": validFollowerID,
				"followee_id": validFollowerID,
			},
			mockReturnErr:  model.ErrCanNotFollowSelf,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Service Error",
			method: http.MethodPost,
			body: map[string]string{
				"follower_id": validFollowerID,
				"followee_id": validFolloweeID,
			},
			mockReturnErr:  errors.New("mock follow error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "Success",
			method: http.MethodPost,
			body: map[string]string{
				"follower_id": validFollowerID,
				"followee_id": validFolloweeID,
			},
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc.ExpectedCalls = nil // reset mock

			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			if m, ok := tt.body.(map[string]string); ok &&
				m["follower_id"] == validFollowerID &&
				m["followee_id"] == validFolloweeID &&
				tt.method == http.MethodPost {
				mockSvc.On("FollowUser", validFolloweeID, validFollowerID).Return(tt.mockReturnErr)
			}

			req := httptest.NewRequest(tt.method, "/follow", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			s.FollowUserHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}
