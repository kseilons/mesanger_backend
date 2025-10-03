package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kseilons/messenger-backend/internal/models"
	"github.com/kseilons/messenger-backend/internal/repository"
)

// UserService interface for user business logic
type UserService interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateStatus(ctx context.Context, userID string, status models.UserStatus) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, limit, offset int) ([]*models.User, error)
	GetOnlineUsers(ctx context.Context) ([]*models.User, error)
}

// userService implements UserService
type userService struct {
	userRepo repository.UserRepository
	logger   *slog.Logger
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, logger *slog.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// Create creates a new user
func (s *userService) Create(ctx context.Context, user *models.User) error {
	// Validate user data
	if user.Username == "" {
		return fmt.Errorf("username is required")
	}
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}

	// Check if username already exists
	existingUser, err := s.userRepo.GetByUsername(ctx, user.Username)
	if err != nil {
		return fmt.Errorf("failed to check username: %w", err)
	}
	if existingUser != nil {
		return fmt.Errorf("username already exists")
	}

	// Check if email already exists
	existingUser, err = s.userRepo.GetByEmail(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("failed to check email: %w", err)
	}
	if existingUser != nil {
		return fmt.Errorf("email already exists")
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User created", "user_id", user.ID, "username", user.Username)
	return nil
}

// GetByID retrieves a user by ID
func (s *userService) GetByID(ctx context.Context, id string) (*models.User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (s *userService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (s *userService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// Update updates a user
func (s *userService) Update(ctx context.Context, user *models.User) error {
	if user.ID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Validate updated data
	if user.Username != "" {
		existingUser, err := s.userRepo.GetByUsername(ctx, user.Username)
		if err != nil {
			return fmt.Errorf("failed to check username: %w", err)
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return fmt.Errorf("username already exists")
		}
	}

	if user.Email != "" {
		existingUser, err := s.userRepo.GetByEmail(ctx, user.Email)
		if err != nil {
			return fmt.Errorf("failed to check email: %w", err)
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return fmt.Errorf("email already exists")
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("User updated", "user_id", user.ID)
	return nil
}

// UpdateStatus updates user status
func (s *userService) UpdateStatus(ctx context.Context, userID string, status models.UserStatus) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	s.logger.Info("User status updated", "user_id", userID, "status", status)
	return nil
}

// Delete deletes a user
func (s *userService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("user ID is required")
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.logger.Info("User deleted", "user_id", id)
	return nil
}

// Search searches for users
func (s *userService) Search(ctx context.Context, query string, limit, offset int) ([]*models.User, error) {
	// Validate parameters
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	users, err := s.userRepo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}

// GetOnlineUsers retrieves all online users
func (s *userService) GetOnlineUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.userRepo.GetOnlineUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get online users: %w", err)
	}

	return users, nil
}
