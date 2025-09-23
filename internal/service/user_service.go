package service

import (
	"context"

	"github.com/kseilons/messenger-backend/internal/logger"
)

type UserService struct {
	logger *logger.Logger
}

func NewUserService(log *logger.Logger) *UserService {
	return &UserService{
		logger: log.WithContext("service", "user"),
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *User) error {
	s.logger.Debug("Creating user", "username", user.Username)

	s.logger.Info("User created successfully", "user_id", user.ID)
	return nil
}
