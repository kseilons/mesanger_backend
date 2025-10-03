package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID          string     `json:"id" db:"id"`
	Username    string     `json:"username" db:"username"`
	Email       string     `json:"email" db:"email"`
	DisplayName string     `json:"display_name" db:"display_name"`
	AvatarURL   string     `json:"avatar_url" db:"avatar_url"`
	Status      UserStatus `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// UserStatus represents user online status
type UserStatus string

const (
	UserStatusOnline  UserStatus = "online"
	UserStatusOffline UserStatus = "offline"
	UserStatusAway    UserStatus = "away"
	UserStatusBusy    UserStatus = "busy"
)
