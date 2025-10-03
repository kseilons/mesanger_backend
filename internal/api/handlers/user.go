package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kseilons/messenger-backend/internal/models"
	"github.com/kseilons/messenger-backend/internal/service"
)

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Status      string `json:"status"`
}

// SearchUsersRequest represents a request to search users
type SearchUsersRequest struct {
	Query  string `form:"q"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

// CreateUser creates a new user
func CreateUser(userService service.UserService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Invalid create user request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user := &models.User{
			Username:    req.Username,
			Email:       req.Email,
			DisplayName: req.DisplayName,
			AvatarURL:   req.AvatarURL,
			Status:      models.UserStatusOffline,
		}

		// TODO: Generate UUID for user ID
		user.ID = "temp-user-id"

		if err := userService.Create(c.Request.Context(), user); err != nil {
			logger.Error("Failed to create user", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		logger.Info("User created", "user_id", user.ID, "username", user.Username)
		c.JSON(http.StatusCreated, user)
	}
}

// GetUser retrieves a user by ID
func GetUser(userService service.UserService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		user, err := userService.GetByID(c.Request.Context(), userID)
		if err != nil {
			logger.Error("Failed to get user", "error", err, "user_id", userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

// UpdateUser updates a user
func UpdateUser(userService service.UserService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		var req UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Invalid update user request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get existing user
		user, err := userService.GetByID(c.Request.Context(), userID)
		if err != nil {
			logger.Error("Failed to get user", "error", err, "user_id", userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Update fields
		if req.Username != "" {
			user.Username = req.Username
		}
		if req.Email != "" {
			user.Email = req.Email
		}
		if req.DisplayName != "" {
			user.DisplayName = req.DisplayName
		}
		if req.AvatarURL != "" {
			user.AvatarURL = req.AvatarURL
		}
		if req.Status != "" {
			user.Status = models.UserStatus(req.Status)
		}

		if err := userService.Update(c.Request.Context(), user); err != nil {
			logger.Error("Failed to update user", "error", err, "user_id", userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		logger.Info("User updated", "user_id", userID)
		c.JSON(http.StatusOK, user)
	}
}

// DeleteUser deletes a user
func DeleteUser(userService service.UserService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		if err := userService.Delete(c.Request.Context(), userID); err != nil {
			logger.Error("Failed to delete user", "error", err, "user_id", userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}

		logger.Info("User deleted", "user_id", userID)
		c.JSON(http.StatusNoContent, nil)
	}
}

// SearchUsers searches for users
func SearchUsers(userService service.UserService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SearchUsersRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			logger.Error("Invalid search users request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set defaults
		if req.Limit <= 0 || req.Limit > 100 {
			req.Limit = 20
		}
		if req.Offset < 0 {
			req.Offset = 0
		}

		users, err := userService.Search(c.Request.Context(), req.Query, req.Limit, req.Offset)
		if err != nil {
			logger.Error("Failed to search users", "error", err, "query", req.Query)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"users":  users,
			"total":  len(users),
			"limit":  req.Limit,
			"offset": req.Offset,
		})
	}
}
