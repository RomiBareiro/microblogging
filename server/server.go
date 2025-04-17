package server

import (
	"context"
	s "microblogging/service"

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
