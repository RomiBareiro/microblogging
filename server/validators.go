package server

import (
	"encoding/json"
	"errors"
	"io"
	m "microblogging/model"
	"net/http"

	"github.com/google/uuid"
)

func ValidateCreatePostInput(body io.Reader) (m.CreatePostRequest, error) {
	var req m.CreatePostRequest
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
