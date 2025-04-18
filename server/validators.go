package server

import (
	"errors"
	m "microblogging/model"
	"net/http"

	"github.com/google/uuid"
)

func ValidateContent(content string) error {
	if len(content) > 280 {
		return errors.New("content too long")
	}
	return nil
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
