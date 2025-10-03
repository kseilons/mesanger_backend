package models

import (
	"time"
)

// Message represents a message in the messenger
type Message struct {
	ID          string      `json:"id" db:"id"`
	GroupID     string      `json:"group_id" db:"group_id"`
	ChannelID   *string     `json:"channel_id" db:"channel_id"`
	SenderID    string      `json:"sender_id" db:"sender_id"`
	Content     string      `json:"content" db:"content"`
	MessageType MessageType `json:"message_type" db:"message_type"`
	ReplyToID   *string     `json:"reply_to_id" db:"reply_to_id"`
	EditedAt    *time.Time  `json:"edited_at" db:"edited_at"`
	DeletedAt   *time.Time  `json:"deleted_at" db:"deleted_at"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`

	// Joined fields for API responses
	Sender      *User               `json:"sender,omitempty"`
	ReplyTo     *Message            `json:"reply_to,omitempty"`
	Reactions   []MessageReaction   `json:"reactions,omitempty"`
	Attachments []MessageAttachment `json:"attachments,omitempty"`
}

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeText    MessageType = "text"
	MessageTypeImage   MessageType = "image"
	MessageTypeFile    MessageType = "file"
	MessageTypeVoice   MessageType = "voice"
	MessageTypeVideo   MessageType = "video"
	MessageTypeSticker MessageType = "sticker"
	MessageTypeSystem  MessageType = "system"
)

// MessageReaction represents a reaction to a message
type MessageReaction struct {
	ID        string    `json:"id" db:"id"`
	MessageID string    `json:"message_id" db:"message_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Emoji     string    `json:"emoji" db:"emoji"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Joined fields
	User *User `json:"user,omitempty"`
}

// MessageAttachment represents an attachment to a message
type MessageAttachment struct {
	ID           string    `json:"id" db:"id"`
	MessageID    string    `json:"message_id" db:"message_id"`
	FileName     string    `json:"file_name" db:"file_name"`
	FileSize     int64     `json:"file_size" db:"file_size"`
	MimeType     string    `json:"mime_type" db:"mime_type"`
	URL          string    `json:"url" db:"url"`
	ThumbnailURL *string   `json:"thumbnail_url" db:"thumbnail_url"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// MessageRead represents when a user read a message
type MessageRead struct {
	ID        string    `json:"id" db:"id"`
	MessageID string    `json:"message_id" db:"message_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	ReadAt    time.Time `json:"read_at" db:"read_at"`
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// WebSocketMessageTypes
const (
	WSMessageTypeNewMessage     = "new_message"
	WSMessageTypeEditMessage    = "edit_message"
	WSMessageTypeDeleteMessage  = "delete_message"
	WSMessageTypeNewReaction    = "new_reaction"
	WSMessageTypeRemoveReaction = "remove_reaction"
	WSMessageTypeUserTyping     = "user_typing"
	WSMessageTypeUserOnline     = "user_online"
	WSMessageTypeUserOffline    = "user_offline"
	WSMessageTypeJoinGroup      = "join_group"
	WSMessageTypeLeaveGroup     = "leave_group"
	WSMessageTypeError          = "error"
)

// TypingStatus represents a user typing status
type TypingStatus struct {
	UserID    string    `json:"user_id"`
	GroupID   string    `json:"group_id"`
	ChannelID *string   `json:"channel_id"`
	IsTyping  bool      `json:"is_typing"`
	Timestamp time.Time `json:"timestamp"`
}
