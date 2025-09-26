package service

import (
	"context"
	"log/slog"

	"github.com/kseilons/messenger-backend/internal/models"
)

type UserService struct {
	logger *slog.Logger
}

func NewUserService(log *slog.Logger) *UserService {
	return &UserService{
		logger: log.With("service", "user"),
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	s.logger.Debug("Creating user", "username", user.Username)

	s.logger.Info("User created successfully", "user_id", user.ID)
	return nil
}
