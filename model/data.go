package model

import (
	"errors"
	"time"
)

type Post struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
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
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

var (
	ErrInvalidUUID         = errors.New("invalid input syntax for type uuid")
	ErrUserNotFound        = errors.New("user not found")
	ErrFolloweeNotFound    = errors.New("followee not found")
	ErrSelfFollow          = errors.New("cannot follow yourself")
	ErrContentTooLong      = errors.New("post content exceeds character limit")
	ErrMissingUserID       = errors.New("user_id is required")
	ErrInvalidJSON         = errors.New("invalid JSON format")
	ErrInvalidRequest      = errors.New("invalid request")
	ErrMethodNotAllowed    = errors.New("method not allowed")
	ErrMissingFollowerID   = errors.New("follower_id is required")
	ErrCouldNotGetTimeline = errors.New("could not get timeline")
	ErrMissingFolloweeID   = errors.New("followee_id is required")
)

type FollowRequest struct {
	FollowerID string `json:"follower_id"`
	FolloweeID string `json:"followee_id"`
}
