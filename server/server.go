package server

import (
	"context"
	"encoding/json"
	"fmt"
	m "microblogging/model"
	"microblogging/service"
	s "microblogging/service"
	"net/http"

	"github.com/gorilla/mux"
)

type server struct {
	Svc s.BlogService
	ctx context.Context
}

func NewServer(ctx context.Context, svc service.BlogService) *server {
	return &server{
		Svc: svc,
		ctx: ctx,
	}
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(m.ErrorResponse{Code: code, Message: message})
}

func (s *server) GetTimelineHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, m.ErrMethodNotAllowed.Error())
		return
	}

	userID := r.URL.Query().Get("user_id")
	if code, message := ValidateUserID(userID); message != "" {
		RespondWithError(w, code, message)
		return
	}
	posts, err := s.Svc.GetTimeline(userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, m.ErrCouldNotGetTimeline.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (s *server) FollowUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, m.ErrMethodNotAllowed.Error())
		return
	}

	req, err := ValidateFollowInput(r.Body)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = s.Svc.FollowUser(req.FollowerID, req.FolloweeID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("failed to follow user %v: %v", req.FolloweeID, err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "followed successfully"})
}
func (s *server) GetFolloweesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, m.ErrMethodNotAllowed.Error())
		return
	}

	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		RespondWithError(w, http.StatusBadRequest, m.ErrMissingUserID.Error())
		return
	}

	followees, err := s.Svc.GetFollowees(userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("failed to fetch followees: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"followees": followees})
}
