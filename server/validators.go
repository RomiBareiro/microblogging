package server

import (
	"errors"

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
