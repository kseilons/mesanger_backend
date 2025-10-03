package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/kseilons/messenger-backend/internal/models"
	"github.com/kseilons/messenger-backend/internal/repository"
)

// MessageService interface for message business logic
type MessageService interface {
	CreateMessage(ctx context.Context, req *CreateMessageRequest) (*models.Message, error)
	GetMessage(ctx context.Context, id string) (*models.Message, error)
	GetMessagesByGroup(ctx context.Context, groupID string, limit, offset int) ([]*models.Message, error)
	GetMessagesByChannel(ctx context.Context, channelID string, limit, offset int) ([]*models.Message, error)
	GetMessageThread(ctx context.Context, messageID string) ([]*models.Message, error)
	UpdateMessage(ctx context.Context, id, content string, userID string) (*models.Message, error)
	DeleteMessage(ctx context.Context, id, userID string) error
	AddReaction(ctx context.Context, messageID, userID, emoji string) (*models.MessageReaction, error)
	RemoveReaction(ctx context.Context, messageID, userID, emoji string) error
	GetReactions(ctx context.Context, messageID string) ([]*models.MessageReaction, error)
	MarkAsRead(ctx context.Context, messageID, userID string) error
	GetUnreadCount(ctx context.Context, userID, groupID string) (int, error)
	AddAttachment(ctx context.Context, messageID, fileName string, fileSize int64, mimeType, url string) (*models.MessageAttachment, error)
	GetAttachments(ctx context.Context, messageID string) ([]*models.MessageAttachment, error)
}

// CreateMessageRequest represents a request to create a message
type CreateMessageRequest struct {
	GroupID     string  `json:"group_id" binding:"required"`
	ChannelID   *string `json:"channel_id"`
	Content     string  `json:"content" binding:"required"`
	MessageType string  `json:"message_type"`
	ReplyToID   *string `json:"reply_to_id"`
}

// messageService implements MessageService
type messageService struct {
	messageRepo repository.MessageRepository
	logger      *slog.Logger
}

// NewMessageService creates a new message service
func NewMessageService(messageRepo repository.MessageRepository, logger *slog.Logger) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		logger:      logger,
	}
}

// CreateMessage creates a new message
func (s *messageService) CreateMessage(ctx context.Context, req *CreateMessageRequest) (*models.Message, error) {
	// Validate message type
	messageType := models.MessageTypeText
	if req.MessageType != "" {
		messageType = models.MessageType(req.MessageType)
		if !isValidMessageType(messageType) {
			return nil, fmt.Errorf("invalid message type: %s", req.MessageType)
		}
	}

	// TODO: Validate user permissions for the group/channel
	// TODO: Get sender ID from context (authenticated user)

	message := &models.Message{
		ID:          uuid.New().String(),
		GroupID:     req.GroupID,
		ChannelID:   req.ChannelID,
		SenderID:    "temp-user-id", // TODO: Get from context
		Content:     req.Content,
		MessageType: messageType,
		ReplyToID:   req.ReplyToID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Get the created message with sender info
	createdMessage, err := s.messageRepo.GetByID(ctx, message.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created message: %w", err)
	}

	s.logger.Info("Message created", "message_id", message.ID, "group_id", req.GroupID)
	return createdMessage, nil
}

// GetMessage retrieves a message by ID
func (s *messageService) GetMessage(ctx context.Context, id string) (*models.Message, error) {
	message, err := s.messageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	if message == nil {
		return nil, fmt.Errorf("message not found")
	}

	return message, nil
}

// GetMessagesByGroup retrieves messages for a group
func (s *messageService) GetMessagesByGroup(ctx context.Context, groupID string, limit, offset int) ([]*models.Message, error) {
	// Validate limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// TODO: Validate user permissions for the group

	messages, err := s.messageRepo.GetByGroup(ctx, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by group: %w", err)
	}

	return messages, nil
}

// GetMessagesByChannel retrieves messages for a channel
func (s *messageService) GetMessagesByChannel(ctx context.Context, channelID string, limit, offset int) ([]*models.Message, error) {
	// Validate limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// TODO: Validate user permissions for the channel

	messages, err := s.messageRepo.GetByChannel(ctx, channelID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by channel: %w", err)
	}

	return messages, nil
}

// GetMessageThread retrieves a message thread (replies)
func (s *messageService) GetMessageThread(ctx context.Context, messageID string) ([]*models.Message, error) {
	// TODO: Validate user permissions

	thread, err := s.messageRepo.GetThread(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message thread: %w", err)
	}

	return thread, nil
}

// UpdateMessage updates a message
func (s *messageService) UpdateMessage(ctx context.Context, id, content string, userID string) (*models.Message, error) {
	// Get the message first
	message, err := s.messageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	if message == nil {
		return nil, fmt.Errorf("message not found")
	}

	// Check if user is the sender
	if message.SenderID != userID {
		return nil, fmt.Errorf("unauthorized: only message sender can edit")
	}

	// Check if message was edited before
	if message.EditedAt != nil {
		return nil, fmt.Errorf("message already edited")
	}

	// Update the message
	message.Content = content
	message.UpdatedAt = time.Now()

	if err := s.messageRepo.Update(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// Get the updated message
	updatedMessage, err := s.messageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated message: %w", err)
	}

	s.logger.Info("Message updated", "message_id", id, "user_id", userID)
	return updatedMessage, nil
}

// DeleteMessage soft deletes a message
func (s *messageService) DeleteMessage(ctx context.Context, id, userID string) error {
	// Get the message first
	message, err := s.messageRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	if message == nil {
		return fmt.Errorf("message not found")
	}

	// Check if user is the sender
	if message.SenderID != userID {
		return fmt.Errorf("unauthorized: only message sender can delete")
	}

	if err := s.messageRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	s.logger.Info("Message deleted", "message_id", id, "user_id", userID)
	return nil
}

// AddReaction adds a reaction to a message
func (s *messageService) AddReaction(ctx context.Context, messageID, userID, emoji string) (*models.MessageReaction, error) {
	// Validate emoji
	if emoji == "" {
		return nil, fmt.Errorf("emoji cannot be empty")
	}

	// Check if message exists
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	if message == nil {
		return nil, fmt.Errorf("message not found")
	}

	// TODO: Validate user permissions for the group/channel

	reaction := &models.MessageReaction{
		ID:        uuid.New().String(),
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
		CreatedAt: time.Now(),
	}

	if err := s.messageRepo.AddReaction(ctx, reaction); err != nil {
		return nil, fmt.Errorf("failed to add reaction: %w", err)
	}

	s.logger.Info("Reaction added", "message_id", messageID, "user_id", userID, "emoji", emoji)
	return reaction, nil
}

// RemoveReaction removes a reaction from a message
func (s *messageService) RemoveReaction(ctx context.Context, messageID, userID, emoji string) error {
	if err := s.messageRepo.RemoveReaction(ctx, messageID, userID, emoji); err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	s.logger.Info("Reaction removed", "message_id", messageID, "user_id", userID, "emoji", emoji)
	return nil
}

// GetReactions retrieves all reactions for a message
func (s *messageService) GetReactions(ctx context.Context, messageID string) ([]*models.MessageReaction, error) {
	reactions, err := s.messageRepo.GetReactions(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reactions: %w", err)
	}

	return reactions, nil
}

// MarkAsRead marks a message as read by a user
func (s *messageService) MarkAsRead(ctx context.Context, messageID, userID string) error {
	if err := s.messageRepo.MarkAsRead(ctx, messageID, userID); err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	return nil
}

// GetUnreadCount gets unread message count for a user in a group
func (s *messageService) GetUnreadCount(ctx context.Context, userID, groupID string) (int, error) {
	count, err := s.messageRepo.GetUnreadCount(ctx, userID, groupID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// AddAttachment adds an attachment to a message
func (s *messageService) AddAttachment(ctx context.Context, messageID, fileName string, fileSize int64, mimeType, url string) (*models.MessageAttachment, error) {
	attachment := &models.MessageAttachment{
		ID:        uuid.New().String(),
		MessageID: messageID,
		FileName:  fileName,
		FileSize:  fileSize,
		MimeType:  mimeType,
		URL:       url,
		CreatedAt: time.Now(),
	}

	if err := s.messageRepo.AddAttachment(ctx, attachment); err != nil {
		return nil, fmt.Errorf("failed to add attachment: %w", err)
	}

	s.logger.Info("Attachment added", "message_id", messageID, "file_name", fileName)
	return attachment, nil
}

// GetAttachments retrieves attachments for a message
func (s *messageService) GetAttachments(ctx context.Context, messageID string) ([]*models.MessageAttachment, error) {
	attachments, err := s.messageRepo.GetAttachments(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachments: %w", err)
	}

	return attachments, nil
}

// isValidMessageType validates message type
func isValidMessageType(messageType models.MessageType) bool {
	validTypes := []models.MessageType{
		models.MessageTypeText,
		models.MessageTypeImage,
		models.MessageTypeFile,
		models.MessageTypeVoice,
		models.MessageTypeVideo,
		models.MessageTypeSticker,
		models.MessageTypeSystem,
	}

	for _, validType := range validTypes {
		if messageType == validType {
			return true
		}
	}

	return false
}
