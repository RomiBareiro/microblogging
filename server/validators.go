package server

import (
	"encoding/json"
	"errors"
	"io"
	m "microblogging/model"
	"net/http"

	"github.com/google/uuid"
)

func ValidateCreatePostInput(body io.Reader) (CreatePostRequest, error) {
	var req CreatePostRequest
	err := json.NewDecoder(body).Decode(&req)
	if err != nil {
		return req, errors.New("invalid input")
	}
	if len(req.Content) > 280 {
		return req, errors.New("content too long")
	}
	return req, nil
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func ValidateUserID(userID string) (int, string) {
	if userID == "" {
		return http.StatusBadRequest, m.ErrMissingUserID.Error()
	}
	if !IsValidUUID(userID) {
		return http.StatusBadRequest, m.ErrInvalidUUID.Error()
	}
	return http.StatusOK, ""
}

func ValidateFollowInput(body io.Reader) (m.FollowRequest, error) {
	var req m.FollowRequest
	err := json.NewDecoder(body).Decode(&req)
	if err != nil || req.FollowerID == "" || req.FolloweeID == "" {
		return req, m.ErrInvalidRequest
	}
	return req, nil
}
