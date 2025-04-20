package server

import (
	"context"
	"encoding/json"
	"microblogging/model"
	s "microblogging/service"
	"net/http"

	"github.com/go-playground/validator"
)

type server struct {
	Svc s.BlogService
	ctx context.Context
}

var validate = validator.New()

func NewServer(ctx context.Context, svc s.BlogService) *server {
	return &server{
		Svc: svc,
		ctx: ctx,
	}
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(model.ErrorResponse{Code: code, Message: message})
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
