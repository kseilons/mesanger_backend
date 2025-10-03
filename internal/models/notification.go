package models

import (
	"time"
)

// Notification represents a notification for a user
type Notification struct {
	ID        string                 `json:"id" db:"id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	Type      NotificationType       `json:"type" db:"type"`
	Title     string                 `json:"title" db:"title"`
	Content   string                 `json:"content" db:"content"`
	Data      map[string]interface{} `json:"data" db:"data"`
	IsRead    bool                   `json:"is_read" db:"is_read"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	ReadAt    *time.Time             `json:"read_at" db:"read_at"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeNewMessage    NotificationType = "new_message"
	NotificationTypeNewReaction   NotificationType = "new_reaction"
	NotificationTypeGroupInvite   NotificationType = "group_invite"
	NotificationTypeGroupUpdate   NotificationType = "group_update"
	NotificationTypeChannelUpdate NotificationType = "channel_update"
	NotificationTypeMention       NotificationType = "mention"
	NotificationTypeSystem        NotificationType = "system"
)

// KafkaEvent represents an event sent to Kafka
type KafkaEvent struct {
	ID        string                 `json:"id"`
	Type      KafkaEventType         `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
}

// KafkaEventType represents the type of Kafka event
type KafkaEventType string

const (
	KafkaEventTypeMessageCreated  KafkaEventType = "message.created"
	KafkaEventTypeMessageEdited   KafkaEventType = "message.edited"
	KafkaEventTypeMessageDeleted  KafkaEventType = "message.deleted"
	KafkaEventTypeReactionAdded   KafkaEventType = "reaction.added"
	KafkaEventTypeReactionRemoved KafkaEventType = "reaction.removed"
	KafkaEventTypeUserJoined      KafkaEventType = "user.joined"
	KafkaEventTypeUserLeft        KafkaEventType = "user.left"
	KafkaEventTypeGroupCreated    KafkaEventType = "group.created"
	KafkaEventTypeGroupUpdated    KafkaEventType = "group.updated"
	KafkaEventTypeChannelCreated  KafkaEventType = "channel.created"
	KafkaEventTypeChannelUpdated  KafkaEventType = "channel.updated"
	KafkaEventTypeUserOnline      KafkaEventType = "user.online"
	KafkaEventTypeUserOffline     KafkaEventType = "user.offline"
)

// EventData represents common event data structures
type EventData struct {
	MessageCreatedData struct {
		Message   *Message `json:"message"`
		GroupID   string   `json:"group_id"`
		ChannelID *string  `json:"channel_id"`
		SenderID  string   `json:"sender_id"`
	} `json:"message_created,omitempty"`

	ReactionData struct {
		MessageID string `json:"message_id"`
		UserID    string `json:"user_id"`
		Emoji     string `json:"emoji"`
		Action    string `json:"action"` // "add" or "remove"
	} `json:"reaction,omitempty"`

	UserStatusData struct {
		UserID  string     `json:"user_id"`
		Status  UserStatus `json:"status"`
		GroupID string     `json:"group_id"`
	} `json:"user_status,omitempty"`

	GroupData struct {
		Group  *Group `json:"group"`
		UserID string `json:"user_id"`
		Action string `json:"action"` // "created", "updated", "joined", "left"
	} `json:"group,omitempty"`
}
