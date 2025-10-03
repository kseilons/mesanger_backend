package models

import (
	"time"
)

// Group represents a group in the messenger
type Group struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        GroupType `json:"type" db:"type"`
	AvatarURL   string    `json:"avatar_url" db:"avatar_url"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// GroupType represents the type of group
type GroupType string

const (
	GroupTypeDirect  GroupType = "direct"  // Direct message between two users
	GroupTypeGroup   GroupType = "group"   // Regular group chat
	GroupTypeChannel GroupType = "channel" // Public channel
)

// GroupMember represents a member of a group
type GroupMember struct {
	ID       string          `json:"id" db:"id"`
	GroupID  string          `json:"group_id" db:"group_id"`
	UserID   string          `json:"user_id" db:"user_id"`
	Role     GroupMemberRole `json:"role" db:"role"`
	JoinedAt time.Time       `json:"joined_at" db:"joined_at"`
}

// GroupMemberRole represents the role of a group member
type GroupMemberRole string

const (
	GroupMemberRoleOwner     GroupMemberRole = "owner"
	GroupMemberRoleAdmin     GroupMemberRole = "admin"
	GroupMemberRoleModerator GroupMemberRole = "moderator"
	GroupMemberRoleMember    GroupMemberRole = "member"
)

// Channel represents a channel within a group
type Channel struct {
	ID          string      `json:"id" db:"id"`
	GroupID     string      `json:"group_id" db:"group_id"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	Type        ChannelType `json:"type" db:"type"`
	IsPrivate   bool        `json:"is_private" db:"is_private"`
	CreatedBy   string      `json:"created_by" db:"created_by"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// ChannelType represents the type of channel
type ChannelType string

const (
	ChannelTypeText  ChannelType = "text"
	ChannelTypeVoice ChannelType = "voice"
	ChannelTypeVideo ChannelType = "video"
)

// ChannelMember represents a member of a channel
type ChannelMember struct {
	ID        string            `json:"id" db:"id"`
	ChannelID string            `json:"channel_id" db:"channel_id"`
	UserID    string            `json:"user_id" db:"user_id"`
	Role      ChannelMemberRole `json:"role" db:"role"`
	JoinedAt  time.Time         `json:"joined_at" db:"joined_at"`
}

// ChannelMemberRole represents the role of a channel member
type ChannelMemberRole string

const (
	ChannelMemberRoleOwner     ChannelMemberRole = "owner"
	ChannelMemberRoleAdmin     ChannelMemberRole = "admin"
	ChannelMemberRoleModerator ChannelMemberRole = "moderator"
	ChannelMemberRoleMember    ChannelMemberRole = "member"
)
