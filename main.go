package main

import (
	"context"
	"microblogging/config"
	"microblogging/service"
)

func main() {
	ctx := context.Background()
	db, err := config.Setup(ctx)
	if err != nil {
		panic(err)
	}
	svc := service.NewBlogService(db)
	config.ServerSetup(svc)
}
