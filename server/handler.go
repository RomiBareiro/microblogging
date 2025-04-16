package server

import (
	"encoding/json"
	"net/http"
)

type CreatePostRequest struct {
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

func (s *server) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := ValidateCreatePostInput(r.Body)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = s.Svc.CreatePost(req.UserID, req.Content)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "could not create post")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "post created"})
}
