package main

import (
	"context"
	"microblogging/config"
	"microblogging/server"
	"microblogging/service"
)

func main() {
	db, err := config.Setup(context.Background())
	if err != nil {
		panic(err)
	}
	svc := service.NewBlogService(db)
	handler := server.NewBlogHandler(svc)

	r := server.SetupRouter(handler)
	r.Run(":8080")
}
