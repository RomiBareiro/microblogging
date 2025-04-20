package service

import (
	"errors"
	"testing"

	"microblogging/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePost(t *testing.T) {
	userID := uuid.New().String()
	content := "Hello, world!"
	expectedUUID := uuid.New()

	tests := map[string]struct {
		setupMock func(mockRepo *MockPostRepository)
		expected  uuid.UUID
		expectErr bool
	}{
		"success": {
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("Save", mock.MatchedBy(func(p *model.Post) bool {
					return p.UserID == userID && p.Content == content
				})).Return(expectedUUID, nil)
			},
			expected:  expectedUUID,
			expectErr: false,
		},
		"repo_error": {
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("Save", mock.Anything).Return(uuid.Nil, errors.New("db error"))
			},
			expected:  uuid.Nil,
			expectErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockRepo := new(MockPostRepository)
			svc := NewBlogService(mockRepo)
			tc.setupMock(mockRepo)

			result, err := svc.CreatePost(userID, content)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetUser(t *testing.T) {
	userID := uuid.New().String()
	expectedUser := model.User{
		ID:   userID,
		Name: "Test User",
	}

	tests := map[string]struct {
		setupMock func(mockRepo *MockPostRepository)
		expected  model.User
		expectErr bool
	}{
		"success": {
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("GetUser", userID).Return(expectedUser, nil)
			},
			expected:  expectedUser,
			expectErr: false,
		},
		"not_found": {
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("GetUser", userID).Return(model.User{}, errors.New("user not found"))
			},
			expected:  model.User{},
			expectErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockRepo := new(MockPostRepository)
			svc := NewBlogService(mockRepo)
			tc.setupMock(mockRepo)

			result, err := svc.GetUser(userID)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
func TestFollowUser(t *testing.T) {
	testCases := []struct {
		name      string
		setupMock func(mockRepo *MockPostRepository)
		input     []string
		expectErr bool
	}{
		{
			name: "success",
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("FollowUser", "user-1", "user-2").Return(nil)
			},
			input:     []string{"user-1", "user-2"},
			expectErr: false,
		},
		{
			name: "db_error",
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("FollowUser", "user-1", "user-2").Return(errors.New("db error"))
			},
			input:     []string{"user-1", "user-2"},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockPostRepository)
			svc := NewBlogService(mockRepo)
			tc.setupMock(mockRepo)

			err := svc.FollowUser(tc.input[0], tc.input[1])

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
func TestGetFollowees(t *testing.T) {
	testCases := []struct {
		name      string
		setupMock func(mockRepo *MockPostRepository)
		input     []string
		limit     int
		expected  []string
		expectErr bool
	}{
		{
			name: "success",
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("GetFollowees", "user-1", 10).Return([]string{"user-2", "user-3"}, nil)
			},
			input:     []string{"user-1"},
			limit:     10,
			expected:  []string{"user-2", "user-3"},
			expectErr: false,
		},
		{
			name: "db_error",
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("GetFollowees", "user-1", 10).Return([]string{}, errors.New("db error"))
			},
			input:     []string{"user-1"},
			limit:     10,
			expected:  []string{},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockPostRepository)
			svc := NewBlogService(mockRepo)
			tc.setupMock(mockRepo)

			result, err := svc.GetFollowees(tc.input[0], tc.limit)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateUser(t *testing.T) {
	testCases := []struct {
		name      string
		setupMock func(mockRepo *MockPostRepository)
		input     model.CreateUserRequest
		expected  uuid.UUID
		expectErr bool
	}{
		{
			name: "success",
			setupMock: func(mockRepo *MockPostRepository) {
				expectedUUID := uuid.Must(uuid.Parse("66e95b4d-1f09-4cfb-b71d-bb80f92a8dbf"))
				mockRepo.On("CreateUser", mock.Anything).Return(expectedUUID, nil)
			},
			input:     model.CreateUserRequest{Name: "John Doe", Email: "oE5W0@example.com", Password: "password123"},
			expected:  uuid.Must(uuid.Parse("66e95b4d-1f09-4cfb-b71d-bb80f92a8dbf")),
			expectErr: false,
		},
		{
			name: "db_error",
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("CreateUser", mock.Anything).Return(uuid.UUID{}, errors.New("db error"))
			},
			input:     model.CreateUserRequest{Name: "Jane Doe"},
			expected:  uuid.UUID{},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockPostRepository)
			svc := NewBlogService(mockRepo)
			tc.setupMock(mockRepo)

			result, err := svc.CreateUser(tc.input)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdatePostPut(t *testing.T) {
	testCases := []struct {
		name      string
		setupMock func(mockRepo *MockPostRepository)
		input     model.CreatePostRequest
		expectErr bool
	}{
		{
			name: "success",
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("UpdatePostPut", model.CreatePostRequest{UserID: "66e95b4d-1f09-4cfb-b71d-bb80f92a8dbf", Content: "Updated content"}).Return(nil)
			},
			input:     model.CreatePostRequest{UserID: "66e95b4d-1f09-4cfb-b71d-bb80f92a8dbf", Content: "Updated content"},
			expectErr: false,
		},
		{
			name: "db_error",
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("UpdatePostPut", model.CreatePostRequest{UserID: "66e95b4d-1f09-4cfb-b71d-bb80f92a8dbf", Content: "Updated content"}).Return(errors.New("db error"))
			},
			input:     model.CreatePostRequest{UserID: "66e95b4d-1f09-4cfb-b71d-bb80f92a8dbf", Content: "Updated content"},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockPostRepository)
			svc := NewBlogService(mockRepo)
			tc.setupMock(mockRepo)

			err := svc.UpdatePostPut(tc.input)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	testCases := []struct {
		name      string
		setupMock func(mockRepo *MockPostRepository)
		input     string
		expectErr bool
	}{
		{
			name: "success",
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("DeleteUser", "user-1").Return(nil)
			},
			input:     "user-1",
			expectErr: false,
		},
		{
			name: "db_error",
			setupMock: func(mockRepo *MockPostRepository) {
				mockRepo.On("DeleteUser", "user-1").Return(errors.New("db error"))
			},
			input:     "user-1",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockPostRepository)
			svc := NewBlogService(mockRepo)
			tc.setupMock(mockRepo)

			err := svc.DeleteUser(tc.input)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
func TestCreatePostSave(t *testing.T) {
	tests := []struct {
		name       string
		input      model.Post
		expectErr  bool
		expectedID uuid.UUID
		mockReturn struct {
			id  uuid.UUID
			err error
		}
	}{
		{
			name: "success",
			input: model.Post{
				Content: "Hello world",
				UserID:  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			},
			expectErr:  false,
			expectedID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			mockReturn: struct {
				id  uuid.UUID
				err error
			}{
				id:  uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
				err: nil,
			},
		},
		{
			name: "repo_error",
			input: model.Post{
				Content: "This should fail",
				UserID:  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			},
			expectErr:  true,
			expectedID: uuid.Nil,
			mockReturn: struct {
				id  uuid.UUID
				err error
			}{
				id:  uuid.Nil,
				err: errors.New("db failure"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPostRepository)

			mockRepo.On("Save", mock.AnythingOfType("*model.Post")).Return(
				tt.mockReturn.id,
				tt.mockReturn.err,
			)

			svc := NewBlogService(mockRepo)

			id, err := svc.CreatePost(tt.input.UserID, tt.input.Content)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, uuid.Nil, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
