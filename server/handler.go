package server

import (
	"encoding/json"
	"fmt"
	m "microblogging/model"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
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

	RespondWithSuccess(w, http.StatusCreated, "post created", map[string]interface{}{
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

	if !IsValidUUID(req.PostID) {
		RespondWithError(w, http.StatusBadRequest, "invalid post_id UUID")
		return
	}
	if err := ValidateContent(req.Content); err != nil {
		RespondWithError(w, http.StatusBadRequest, m.ErrContentTooLong.Error())
		return
	}
	err := s.Svc.UpdatePostPut(req)
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

	query := r.URL.Query()
	userID := query.Get("user_id")
	limitStr := query.Get("limit")
	beforeStr := query.Get("before")

	req, err := loadTimelineParams(userID, limitStr, beforeStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	posts, err := s.Svc.GetTimeline(req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, m.ErrCouldNotGetTimeline.Error())
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Timeline info", map[string]interface{}{
		"user_id": req.UserID,
		"posts":   posts,
	})
}

func loadTimelineParams(userID, limitStr, beforeStr string) (m.TimelineRequest, error) {
	var (
		r        m.TimelineRequest
		errGroup errgroup.Group
		mu       sync.Mutex
	)

	errGroup.Go(func() error {
		if !IsValidUUID(userID) {
			return m.ErrInvalidUUID
		}
		mu.Lock()
		r.UserID = userID
		mu.Unlock()
		return nil
	})

	errGroup.Go(func() error {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			limit = 50
		}
		mu.Lock()
		r.Limit = limit
		mu.Unlock()
		return nil
	})

	errGroup.Go(func() error {
		var before time.Time
		if beforeStr != "" {
			var err error
			before, err = time.Parse(time.RFC3339, beforeStr)
			if err != nil {
				return fmt.Errorf("invalid before parameter: %v", err)
			}
		} else {
			before = time.Now().AddDate(0, 0, -3)
		}
		mu.Lock()
		r.Before = before
		mu.Unlock()
		return nil
	})

	// wait for all goroutines to finish
	if err := errGroup.Wait(); err != nil {
		return m.TimelineRequest{}, err
	}

	return r, nil
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

	if req.FollowerID == req.FolloweeID {
		RespondWithError(w, http.StatusBadRequest, m.ErrCanNotFollowSelf.Error())
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

func (s *server) UnfollowUserHandler(w http.ResponseWriter, r *http.Request) {
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

	if req.FollowerID == req.FolloweeID {
		RespondWithError(w, http.StatusBadRequest, m.ErrCanNotUnfollowSelf.Error())
		return
	}

	err := s.Svc.UnfollowUser(req.FollowerID, req.FolloweeID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("failed to unfollow user %v: %v", req.FolloweeID, err))
		return
	}

	RespondWithSuccess(w, http.StatusOK, "user unfollowed", map[string]interface{}{
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
