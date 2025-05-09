package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" validate:"required,uuid" db:"user_id" `
	Content   string    `json:"content" validate:"required,max=280" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"created_at"`
}
type Follow struct {
	FollowerID string
	FolloweeID string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CreatePostRequest struct {
	UserID  string `json:"user_id" validate:"required,uuid"`
	Content string `json:"content"`
	PostID  string `json:"post_id"`
}

var (
	ErrInvalidUUID         = errors.New("invalid input syntax for type uuid")
	ErrUserNotFound        = errors.New("user not found")
	ErrFolloweeNotFound    = errors.New("followee not found")
	ErrCanNotFollowSelf    = errors.New("can not follow yourself")
	ErrCanNotUnfollowSelf  = errors.New("can not unfollow yourself")
	ErrContentTooLong      = errors.New("post content exceeds character limit")
	ErrMissingUserID       = errors.New("user_id is required")
	ErrInvalidJSON         = errors.New("invalid JSON format")
	ErrInvalidRequest      = errors.New("invalid request")
	ErrMethodNotAllowed    = errors.New("method not allowed")
	ErrMissingFollowerID   = errors.New("follower_id is required")
	ErrCouldNotGetTimeline = errors.New("could not get timeline")
	ErrCouldNotCreateUser  = errors.New("could not create user")
	ErrMissingFolloweeID   = errors.New("followee_id is required")
	ErrInvalidInput        = errors.New("invalid input")
	ErrCouldNotUpdate      = errors.New("could not update post")
	ErrPostNotFound        = errors.New("post not found")
)

type FollowRequest struct {
	FollowerID string `json:"follower_id" validate:"required,uuid" db:"follower_id"`
	FolloweeID string `json:"followee_id" validate:"required,uuid" db:"followee_id"`
}

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type User struct {
	ID         string    `json:"id" db:"id"`
	Name       string    `json:"name" db:"user_name"`
	LastPostID uuid.UUID `json:"last_post_id" db:"last_post_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type TimelineRequest struct {
	UserID string    `json:"user_id"`
	Limit  int       `json:"limit"`
	Before time.Time `json:"before"`
}

type TimelineResponse struct {
	Posts []Post `json:"posts"`
}
