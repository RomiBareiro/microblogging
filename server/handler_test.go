package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"microblogging/model"
	"microblogging/server"
	"microblogging/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreatePostHandler(t *testing.T) {
	mockSvc := &service.MockBlogService{}
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
	mockSvc := &service.MockBlogService{}
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
	mockSvc := &service.MockBlogService{}
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
				mockSvc.On("UnfollowUser", validFollowerID, validFolloweeID).Return(tt.mockReturnErr)
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
	mockSvc := &service.MockBlogService{}
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
				mockSvc.On("FollowUser", validFollowerID, validFolloweeID).Return(tt.mockReturnErr)
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
