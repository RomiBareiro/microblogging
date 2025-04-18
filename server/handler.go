package server

import (
	"encoding/json"
	"fmt"
	m "microblogging/model"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (s *server) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, m.ErrMethodNotAllowed.Error())
		return
	}

	req, err := ValidateCreatePostInput(r.Body)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := s.Svc.CreatePost(req.UserID, req.Content)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("could not create post: %v", err))
		return
	}

	RespondWithSuccess(w, http.StatusCreated, "post created/updated", map[string]interface{}{
		"user_id": req.UserID,
		"post_id": id,
	})
}

func (s *server) CreatePutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		RespondWithError(w, http.StatusMethodNotAllowed, m.ErrMethodNotAllowed.Error())
		return
	} //terminar
}

func (s *server) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, m.ErrMethodNotAllowed.Error())
		return
	}

	var req m.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, m.ErrInvalidRequest.Error())
		return
	}

	if err := validate.Struct(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, m.ErrInvalidJSON.Error())
		return
	}

	user, err := s.Svc.CreateUser(req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, m.ErrCouldNotCreateUser.Error())
		return
	}

	RespondWithSuccess(w, http.StatusCreated, "user created", map[string]interface{}{
		"user_id": user,
	})
}

func (s *server) GetTimelineHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, m.ErrMethodNotAllowed.Error())
		return
	}

	var req m.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	posts, err := s.Svc.GetTimeline(req.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, m.ErrCouldNotGetTimeline.Error())
		return
	}

	RespondWithSuccess(w, http.StatusOK, "posts", map[string]interface{}{
		"user_id": req.UserID,
		"posts":   posts,
	})
}

func (s *server) FollowUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, m.ErrMethodNotAllowed.Error())
		return
	}

	var req m.FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, m.ErrInvalidRequest.Error())
		return
	}

	if err := validate.Struct(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := s.Svc.FollowUser(req.FollowerID, req.FolloweeID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("failed to follow user %v: %v", req.FolloweeID, err))
		return
	}

	RespondWithSuccess(w, http.StatusOK, "user followed", map[string]interface{}{
		"user_id":  req.FollowerID,
		"followee": req.FolloweeID,
	})
}

func (s *server) GetFolloweesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, m.ErrMethodNotAllowed.Error())
		return
	}

	vars := mux.Vars(r)
	userID := vars["id"]
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid limit parameter")
		return
	}
	if userID == "" {
		RespondWithError(w, http.StatusBadRequest, m.ErrMissingUserID.Error())
		return
	}
	if limit <= 0 {
		RespondWithError(w, http.StatusBadRequest, "limit must be greater than 0")
		return
	}
	followees, err := s.Svc.GetFollowees(userID, limit)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("failed to fetch followees: %v", err))
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Got user followees", map[string]interface{}{
		"user_id":   userID,
		"followees": followees,
	})
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(m.ErrorResponse{Code: code, Message: message})
}

func RespondWithSuccess(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := struct {
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}{
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}
