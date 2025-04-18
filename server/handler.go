package server

import (
	"encoding/json"
	"fmt"
	m "microblogging/model"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (s *server) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, m.ErrMethodNotAllowed.Error())
		return
	}
	var req m.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, m.ErrInvalidRequest.Error())
		return
	}

	if err := validate.Struct(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, m.ErrInvalidJSON.Error())
		return
	}
	if err := ValidateContent(req.Content); err != nil {
		RespondWithError(w, http.StatusBadRequest, m.ErrContentTooLong.Error())
		return
	}

	id, err := s.Svc.CreatePost(req.UserID, req.Content)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("could not create post: %v", err))
		return
	}

	RespondWithSuccess(w, http.StatusCreated, "post createdd", map[string]interface{}{
		"user_id": req.UserID,
		"post_id": id,
	})
}

func (s *server) UpdatePostPutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		RespondWithError(w, http.StatusMethodNotAllowed, m.ErrMethodNotAllowed.Error())
		return
	}
	var req m.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, m.ErrInvalidJSON.Error())
		return
	}

	_, err := uuid.Parse(req.PostID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid post_id UUID")
		return
	}
	if err := ValidateContent(req.Content); err != nil {
		RespondWithError(w, http.StatusBadRequest, m.ErrContentTooLong.Error())
		return
	}
	err = s.Svc.UpdatePostPut(req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s: %s", m.ErrCouldNotUpdate.Error(), err.Error()))
		return
	}

	RespondWithSuccess(w, http.StatusOK, "post updated", map[string]interface{}{
		"user_id": req.UserID,
		"post_id": req.PostID,
	})
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

func (s *server) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		RespondWithError(w, http.StatusMethodNotAllowed, m.ErrMethodNotAllowed.Error())
		return
	}

	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		RespondWithError(w, http.StatusBadRequest, m.ErrMissingUserID.Error())
		return
	}
	if !IsValidUUID(userID) {
		RespondWithError(w, http.StatusBadRequest, m.ErrInvalidUUID.Error())
		return
	}
	err := s.Svc.DeleteUser(userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("failed to delete user: %v", err))
		return
	}
	RespondWithSuccess(w, http.StatusOK, "user deleted", map[string]interface{}{
		"user_id": userID,
	})
}
