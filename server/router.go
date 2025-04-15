package server

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes the Gin engine with defined routes and handlers.
// It sets up POST and GET endpoints for creating posts and retrieving timelines,
// respectively, using the provided BlogHandler.

func SetupRouter(h *BlogHandler) *gin.Engine {
	r := gin.Default()

	r.POST("/posts", h.CreatePost)
	r.GET("/timeline", h.GetTimeline)
	r.POST("/follow", h.FollowUser)
	r.GET("/followees/:id", h.GetFollowees)

	return r
}
