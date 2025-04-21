package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"microblogging/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSave(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(sqlmock.Sqlmock)
		inputPost    *model.Post
		expectedErr  bool
		expectedUUID bool // true if we expect a valid uuid
	}{
		{
			name: "Success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`
				INSERT INTO posts (user_id, content, created_at, updated_at)
				VALUES ($1, $2, $3, $4)
				RETURNING id;
			`)).
					WithArgs("user-id-123", "Hello world", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			inputPost: &model.Post{
				UserID:  "user-id-123",
				Content: "Hello world",
			},
			expectedErr:  false,
			expectedUUID: true,
		},
		{
			name: "DB error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`
					INSERT INTO posts (user_id, content, created_at, updated_at)
					VALUES ($1, $2, $3, $4)
					RETURNING id;
				`)).
					WithArgs("user-id-123", "Hello world", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			inputPost: &model.Post{
				UserID:  "user-id-123",
				Content: "Hello world",
			},
			expectedErr:  true,
			expectedUUID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			logger := zap.NewNop()
			sqlxDB := sqlx.NewDb(db, "sqlmock")
			repo := &DBConnector{DB: sqlxDB, Logger: logger}

			tt.setupMock(mock)

			id, err := repo.Save(tt.inputPost)
			if tt.expectedErr {
				assert.Error(t, err)
				assert.Equal(t, uuid.Nil, id)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, id)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdatePostPut(t *testing.T) {
	validUUID := uuid.New()
	userID := "user-id-123"
	content := "Updated post content"

	tests := []struct {
		name        string
		input       model.CreatePostRequest
		setupMock   func(sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name: "Success",
			input: model.CreatePostRequest{
				PostID:  validUUID.String(),
				UserID:  userID,
				Content: content,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock existPost query
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1 AND user_id = $2)
				`)).
					WithArgs(validUUID, userID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

				// Mock update query
				mock.ExpectExec(regexp.QuoteMeta(`
					UPDATE posts
					SET content = $1, updated_at = $2
					WHERE id = $3 AND user_id = $4;
				`)).
					WithArgs(content, sqlmock.AnyArg(), validUUID.String(), userID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedErr: nil,
		},
		{
			name: "Post not found",
			input: model.CreatePostRequest{
				PostID:  validUUID.String(),
				UserID:  userID,
				Content: content,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock existPost query returns false
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1 AND user_id = $2)
				`)).
					WithArgs(validUUID, userID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			expectedErr: model.ErrPostNotFound,
		},
		{
			name: "Invalid UUID",
			input: model.CreatePostRequest{
				PostID:  "not-a-uuid",
				UserID:  userID,
				Content: content,
			},
			setupMock:   func(mock sqlmock.Sqlmock) {}, // no DB interaction
			expectedErr: model.ErrInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			logger := zap.NewNop()
			sqlxDB := sqlx.NewDb(db, "sqlmock")
			repo := &DBConnector{DB: sqlxDB, Logger: logger}

			tt.setupMock(mock)

			err = repo.UpdatePostPut(tt.input)
			assert.Equal(t, tt.expectedErr, err)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetTimeline(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := zap.NewNop()
	repo := &DBConnector{DB: sqlxDB, Logger: logger}
	now := time.Now()
	mock.ExpectQuery(`SELECT p.id, p.user_id, p.content, p.created_at FROM posts`).
		WithArgs("user-id-123", now, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "created_at"}).
			AddRow(uuid.New(), "user-id-123", "Hello!", now))

	timeline, err := repo.GetTimeline(model.TimelineRequest{
		UserID: "user-id-123",
		Before: now,
		Limit:  10,
	})

	assert.NoError(t, err)
	assert.Len(t, timeline.Posts, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowUserCases(t *testing.T) {
	type args struct {
		follower string
		followee string
	}
	tests := []struct {
		name      string
		args      args
		expectErr bool
	}{
		{
			name:      "success",
			args:      args{"user1", "user2"},
			expectErr: false,
		},
		{
			name:      "followee_not_found",
			args:      args{"user1", "user2"},
			expectErr: true,
		},
		{
			name:      "follower_not_found",
			args:      args{"user1", "user2"},
			expectErr: true,
		},
		{
			name:      "insert_error",
			args:      args{"user1", "user2"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")
			logger := zap.NewNop()
			repo := &DBConnector{DB: sqlxDB, Logger: logger}

			switch tt.name {
			case "success":
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.followee).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.follower).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectExec(`INSERT INTO follows`).
					WithArgs(tt.args.follower, tt.args.followee).
					WillReturnResult(sqlmock.NewResult(1, 1))

			case "followee_not_found":
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.followee).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

			case "follower_not_found":
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.followee).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.follower).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

			case "insert_error":
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.followee).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.follower).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectExec(`INSERT INTO follows`).
					WithArgs(tt.args.follower, tt.args.followee).
					WillReturnError(errors.New("insert failed"))
			}

			err = repo.FollowUser(tt.args.follower, tt.args.followee)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUnfollowUserCases(t *testing.T) {
	type args struct {
		follower string
		followee string
	}
	tests := []struct {
		name      string
		args      args
		expectErr bool
	}{
		{name: "success", args: args{"user1", "user2"}, expectErr: false},
		{name: "followee_not_found", args: args{"user1", "user2"}, expectErr: true},
		{name: "follower_not_found", args: args{"user1", "user2"}, expectErr: true},
		{name: "update_error", args: args{"user1", "user2"}, expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			logger := zap.NewNop()
			repo := &DBConnector{DB: sqlxDB, Logger: logger}

			switch tt.name {
			case "success":
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.followee).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.follower).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectExec(`UPDATE follows\s+SET is_active = FALSE\s+WHERE follower_id = \$1 AND followee_id = \$2;?`).
					WithArgs(tt.args.follower, tt.args.followee).
					WillReturnResult(sqlmock.NewResult(0, 1))

			case "followee_not_found":
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.followee).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

			case "follower_not_found":
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.followee).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.follower).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

			case "update_error":
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.followee).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM users WHERE id = \$1\s*\)`).
					WithArgs(tt.args.follower).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectExec(`UPDATE follows\s+SET is_active = FALSE\s+WHERE follower_id = \$1 AND followee_id = \$2;?`).
					WithArgs(tt.args.follower, tt.args.followee).
					WillReturnError(fmt.Errorf("update failed"))
			}

			err = repo.UnfollowUser(tt.args.follower, tt.args.followee)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetFollowees(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := zap.NewNop()
	repo := &DBConnector{DB: sqlxDB, Logger: logger}

	mock.ExpectQuery(`SELECT followee_id FROM follows`).
		WithArgs("user1", 5).
		WillReturnRows(sqlmock.NewRows([]string{"followee_id"}).AddRow("user2").AddRow("user3"))

	followees, err := repo.GetFollowees("user1", 5)

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"user2", "user3"}, followees)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser(t *testing.T) {
	// Set up the mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	fixedTime := time.Date(2025, 4, 18, 23, 12, 40, 0, time.UTC).Format(time.RFC3339)

	// Create the repository with the mock database
	r := &DBConnector{
		DB:     sqlxDB,
		Logger: zap.NewNop(),
	}

	// Input data for creating the user
	userData := model.CreateUserRequest{
		Name:     "testuser",
		Password: "securepassword",
		Email:    "test@example.com",
	}

	// Happy Path Test (user created successfully)
	t.Run("happy_path_-_user_created_successfully", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO users \(.+\)`).
			WithArgs(userData.Name, userData.Password, userData.Email, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))

		userID, err := r.CreateUser(userData)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, userID)
	})

	// Test case for database error when inserting the user
	t.Run("db_returns_error_on_insert", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO users \(.+\)`).
			WithArgs(userData.Name, userData.Password, userData.Email, fixedTime, fixedTime).
			WillReturnError(fmt.Errorf("db error"))

		userID, err := r.CreateUser(userData)

		// Verify the results
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, userID)
	})

	// Test case where no ID is returned from the database
	t.Run("no_rows_returned", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO users \(.+\)`).
			WithArgs(userData.Name, userData.Password, userData.Email, fixedTime, fixedTime).
			WillReturnRows(sqlmock.NewRows([]string{"id"})) // No ID returned

		userID, err := r.CreateUser(userData)

		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, userID)
	})
}

func TestGetUser(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := zap.NewNop()
	repo := &DBConnector{DB: sqlxDB, Logger: logger}
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM users WHERE id = \$1`).
		WithArgs("user-id-123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_name", "created_at", "updated_at"}).
			AddRow("user-id-123", "alice", now, now))

	user, err := repo.GetUser("user-id-123")

	assert.NoError(t, err)
	assert.Equal(t, "alice", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUser(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := zap.NewNop()
	repo := &DBConnector{DB: sqlxDB, Logger: logger}

	mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM users WHERE id = \$1\)`).
		WithArgs("user-id-123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs("user-id-123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.DeleteUser("user-id-123")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
