package main

import (
	"microblogging/config"
	"microblogging/repository"
	"microblogging/server"
	"microblogging/service"
)

func main() {
	db := config.ConnectDB()
	repo := repository.NewPostRepository(db)
	svc := service.NewBlogService(repo)
	handler := server.NewBlogHandler(svc)

	r := server.SetupRouter(handler)
	r.Run(":8080")
}
