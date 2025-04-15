package server

import (
	"microblogging/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BlogHandler struct {
	service service.BlogService
}

func NewBlogHandler(s service.BlogService) *BlogHandler {
	return &BlogHandler{service: s}
}

func (h *BlogHandler) CreatePost(c *gin.Context) {
	var req struct {
		UserID  string `json:"user_id"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if len(req.Content) > 280 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content too long"})
		return
	}

	if err := h.service.CreatePost(req.UserID, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create post"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "post created"})
}

func (h *BlogHandler) GetTimeline(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	posts, err := h.service.GetTimeline(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get timeline"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

type FollowRequest struct {
	FollowerID string `json:"follower_id" binding:"required"`
	FolloweeID string `json:"followee_id" binding:"required"`
}

// FollowUser handles the HTTP POST request to follow a user.
// It expects the user to be logged in and the request body to contain the IDs of the follower and followee.
// The request is expected to be in JSON format.
func (h *BlogHandler) FollowUser(c *gin.Context) {
	var req FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.FollowUser(req.FollowerID, req.FolloweeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Followed successfully"})
}

// GetFollowees handles the HTTP GET request to retrieve the list of followees
// for a specified user. It expects the user ID to be provided as a URL parameter.
func (h *BlogHandler) GetFollowees(c *gin.Context) {
	userID := c.Param("id")

	followees, err := h.service.GetFollowees(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch followees"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"followees": followees})
}
