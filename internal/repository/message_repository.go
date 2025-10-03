package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/kseilons/messenger-backend/internal/models"
)

// MessageRepository interface for message data operations
type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	GetByID(ctx context.Context, id string) (*models.Message, error)
	GetByGroup(ctx context.Context, groupID string, limit, offset int) ([]*models.Message, error)
	GetByChannel(ctx context.Context, channelID string, limit, offset int) ([]*models.Message, error)
	GetThread(ctx context.Context, messageID string) ([]*models.Message, error)
	Update(ctx context.Context, message *models.Message) error
	Delete(ctx context.Context, id string) error
	AddReaction(ctx context.Context, reaction *models.MessageReaction) error
	RemoveReaction(ctx context.Context, messageID, userID, emoji string) error
	GetReactions(ctx context.Context, messageID string) ([]*models.MessageReaction, error)
	MarkAsRead(ctx context.Context, messageID, userID string) error
	GetUnreadCount(ctx context.Context, userID, groupID string) (int, error)
	AddAttachment(ctx context.Context, attachment *models.MessageAttachment) error
	GetAttachments(ctx context.Context, messageID string) ([]*models.MessageAttachment, error)
}

// messageRepository implements MessageRepository
type messageRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sql.DB, logger *slog.Logger) MessageRepository {
	return &messageRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new message
func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	query := `
		INSERT INTO messages (id, group_id, channel_id, sender_id, content, message_type, reply_to_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	var channelID interface{}
	if message.ChannelID != nil {
		channelID = *message.ChannelID
	}

	var replyToID interface{}
	if message.ReplyToID != nil {
		replyToID = *message.ReplyToID
	}

	_, err := r.db.ExecContext(ctx, query,
		message.ID, message.GroupID, channelID, message.SenderID,
		message.Content, message.MessageType, replyToID)

	if err != nil {
		r.logger.Error("Failed to create message", "error", err, "message_id", message.ID)
		return fmt.Errorf("failed to create message: %w", err)
	}

	r.logger.Info("Message created", "message_id", message.ID, "group_id", message.GroupID)
	return nil
}

// GetByID retrieves a message by ID
func (r *messageRepository) GetByID(ctx context.Context, id string) (*models.Message, error) {
	query := `
		SELECT m.id, m.group_id, m.channel_id, m.sender_id, m.content, m.message_type, 
		       m.reply_to_id, m.edited_at, m.deleted_at, m.created_at, m.updated_at,
		       u.id, u.username, u.display_name, u.avatar_url, u.status
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		WHERE m.id = $1 AND m.deleted_at IS NULL
	`

	message := &models.Message{}
	var channelID, replyToID sql.NullString
	var editedAt, deletedAt sql.NullTime
	sender := &models.User{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID, &message.GroupID, &channelID, &message.SenderID,
		&message.Content, &message.MessageType, &replyToID,
		&editedAt, &deletedAt, &message.CreatedAt, &message.UpdatedAt,
		&sender.ID, &sender.Username, &sender.DisplayName, &sender.AvatarURL, &sender.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get message by ID", "error", err, "message_id", id)
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}

	if channelID.Valid {
		message.ChannelID = &channelID.String
	}
	if replyToID.Valid {
		message.ReplyToID = &replyToID.String
	}
	if editedAt.Valid {
		message.EditedAt = &editedAt.Time
	}
	if deletedAt.Valid {
		message.DeletedAt = &deletedAt.Time
	}

	message.Sender = sender

	return message, nil
}

// GetByGroup retrieves messages by group ID
func (r *messageRepository) GetByGroup(ctx context.Context, groupID string, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT m.id, m.group_id, m.channel_id, m.sender_id, m.content, m.message_type,
		       m.reply_to_id, m.edited_at, m.deleted_at, m.created_at, m.updated_at,
		       u.id, u.username, u.display_name, u.avatar_url, u.status
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		WHERE m.group_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, groupID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get messages by group", "error", err, "group_id", groupID)
		return nil, fmt.Errorf("failed to get messages by group: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// GetByChannel retrieves messages by channel ID
func (r *messageRepository) GetByChannel(ctx context.Context, channelID string, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT m.id, m.group_id, m.channel_id, m.sender_id, m.content, m.message_type,
		       m.reply_to_id, m.edited_at, m.deleted_at, m.created_at, m.updated_at,
		       u.id, u.username, u.display_name, u.avatar_url, u.status
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		WHERE m.channel_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, channelID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get messages by channel", "error", err, "channel_id", channelID)
		return nil, fmt.Errorf("failed to get messages by channel: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// GetThread retrieves message thread (replies)
func (r *messageRepository) GetThread(ctx context.Context, messageID string) ([]*models.Message, error) {
	query := `SELECT * FROM get_message_thread($1)`

	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		r.logger.Error("Failed to get message thread", "error", err, "message_id", messageID)
		return nil, fmt.Errorf("failed to get message thread: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		message := &models.Message{}
		var channelID, replyToID sql.NullString

		err := rows.Scan(
			&message.ID, &message.GroupID, &channelID, &message.SenderID,
			&message.Content, &message.MessageType, &replyToID, &message.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan thread message", "error", err)
			return nil, fmt.Errorf("failed to scan thread message: %w", err)
		}

		if channelID.Valid {
			message.ChannelID = &channelID.String
		}
		if replyToID.Valid {
			message.ReplyToID = &replyToID.String
		}

		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate thread messages: %w", err)
	}

	return messages, nil
}

// Update updates a message
func (r *messageRepository) Update(ctx context.Context, message *models.Message) error {
	query := `
		UPDATE messages
		SET content = $2, edited_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, message.ID, message.Content)
	if err != nil {
		r.logger.Error("Failed to update message", "error", err, "message_id", message.ID)
		return fmt.Errorf("failed to update message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	r.logger.Info("Message updated", "message_id", message.ID)
	return nil
}

// Delete soft deletes a message
func (r *messageRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE messages
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete message", "error", err, "message_id", id)
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	r.logger.Info("Message deleted", "message_id", id)
	return nil
}

// AddReaction adds a reaction to a message
func (r *messageRepository) AddReaction(ctx context.Context, reaction *models.MessageReaction) error {
	query := `
		INSERT INTO message_reactions (id, message_id, user_id, emoji)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (message_id, user_id, emoji) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, reaction.ID, reaction.MessageID, reaction.UserID, reaction.Emoji)
	if err != nil {
		r.logger.Error("Failed to add reaction", "error", err, "message_id", reaction.MessageID)
		return fmt.Errorf("failed to add reaction: %w", err)
	}

	r.logger.Info("Reaction added", "message_id", reaction.MessageID, "emoji", reaction.Emoji)
	return nil
}

// RemoveReaction removes a reaction from a message
func (r *messageRepository) RemoveReaction(ctx context.Context, messageID, userID, emoji string) error {
	query := `
		DELETE FROM message_reactions
		WHERE message_id = $1 AND user_id = $2 AND emoji = $3
	`

	result, err := r.db.ExecContext(ctx, query, messageID, userID, emoji)
	if err != nil {
		r.logger.Error("Failed to remove reaction", "error", err, "message_id", messageID)
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("reaction not found")
	}

	r.logger.Info("Reaction removed", "message_id", messageID, "emoji", emoji)
	return nil
}

// GetReactions retrieves all reactions for a message
func (r *messageRepository) GetReactions(ctx context.Context, messageID string) ([]*models.MessageReaction, error) {
	query := `
		SELECT mr.id, mr.message_id, mr.user_id, mr.emoji, mr.created_at,
		       u.id, u.username, u.display_name, u.avatar_url
		FROM message_reactions mr
		LEFT JOIN users u ON mr.user_id = u.id
		WHERE mr.message_id = $1
		ORDER BY mr.created_at
	`

	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		r.logger.Error("Failed to get reactions", "error", err, "message_id", messageID)
		return nil, fmt.Errorf("failed to get reactions: %w", err)
	}
	defer rows.Close()

	var reactions []*models.MessageReaction
	for rows.Next() {
		reaction := &models.MessageReaction{}
		user := &models.User{}

		err := rows.Scan(
			&reaction.ID, &reaction.MessageID, &reaction.UserID, &reaction.Emoji, &reaction.CreatedAt,
			&user.ID, &user.Username, &user.DisplayName, &user.AvatarURL,
		)
		if err != nil {
			r.logger.Error("Failed to scan reaction", "error", err)
			return nil, fmt.Errorf("failed to scan reaction: %w", err)
		}

		reaction.User = user
		reactions = append(reactions, reaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate reactions: %w", err)
	}

	return reactions, nil
}

// MarkAsRead marks a message as read by a user
func (r *messageRepository) MarkAsRead(ctx context.Context, messageID, userID string) error {
	query := `
		INSERT INTO message_reads (id, message_id, user_id, read_at)
		VALUES (gen_random_uuid(), $1, $2, NOW())
		ON CONFLICT (message_id, user_id) DO UPDATE SET read_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query, messageID, userID)
	if err != nil {
		r.logger.Error("Failed to mark message as read", "error", err, "message_id", messageID, "user_id", userID)
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	return nil
}

// GetUnreadCount gets unread message count for a user in a group
func (r *messageRepository) GetUnreadCount(ctx context.Context, userID, groupID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages m
		LEFT JOIN message_reads mr ON m.id = mr.message_id AND mr.user_id = $1
		WHERE m.group_id = $2 
		AND m.deleted_at IS NULL 
		AND m.sender_id != $1
		AND mr.id IS NULL
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID, groupID).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to get unread count", "error", err, "user_id", userID, "group_id", groupID)
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// AddAttachment adds an attachment to a message
func (r *messageRepository) AddAttachment(ctx context.Context, attachment *models.MessageAttachment) error {
	query := `
		INSERT INTO message_attachments (id, message_id, file_name, file_size, mime_type, url, thumbnail_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	var thumbnailURL interface{}
	if attachment.ThumbnailURL != nil {
		thumbnailURL = *attachment.ThumbnailURL
	}

	_, err := r.db.ExecContext(ctx, query,
		attachment.ID, attachment.MessageID, attachment.FileName,
		attachment.FileSize, attachment.MimeType, attachment.URL, thumbnailURL)

	if err != nil {
		r.logger.Error("Failed to add attachment", "error", err, "message_id", attachment.MessageID)
		return fmt.Errorf("failed to add attachment: %w", err)
	}

	r.logger.Info("Attachment added", "message_id", attachment.MessageID, "file_name", attachment.FileName)
	return nil
}

// GetAttachments retrieves attachments for a message
func (r *messageRepository) GetAttachments(ctx context.Context, messageID string) ([]*models.MessageAttachment, error) {
	query := `
		SELECT id, message_id, file_name, file_size, mime_type, url, thumbnail_url, created_at
		FROM message_attachments
		WHERE message_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		r.logger.Error("Failed to get attachments", "error", err, "message_id", messageID)
		return nil, fmt.Errorf("failed to get attachments: %w", err)
	}
	defer rows.Close()

	var attachments []*models.MessageAttachment
	for rows.Next() {
		attachment := &models.MessageAttachment{}
		var thumbnailURL sql.NullString

		err := rows.Scan(
			&attachment.ID, &attachment.MessageID, &attachment.FileName,
			&attachment.FileSize, &attachment.MimeType, &attachment.URL,
			&thumbnailURL, &attachment.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan attachment", "error", err)
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}

		if thumbnailURL.Valid {
			attachment.ThumbnailURL = &thumbnailURL.String
		}

		attachments = append(attachments, attachment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate attachments: %w", err)
	}

	return attachments, nil
}

// scanMessages scans message rows from database
func (r *messageRepository) scanMessages(rows *sql.Rows) ([]*models.Message, error) {
	var messages []*models.Message
	for rows.Next() {
		message := &models.Message{}
		var channelID, replyToID sql.NullString
		var editedAt, deletedAt sql.NullTime
		sender := &models.User{}

		err := rows.Scan(
			&message.ID, &message.GroupID, &channelID, &message.SenderID,
			&message.Content, &message.MessageType, &replyToID,
			&editedAt, &deletedAt, &message.CreatedAt, &message.UpdatedAt,
			&sender.ID, &sender.Username, &sender.DisplayName, &sender.AvatarURL, &sender.Status,
		)
		if err != nil {
			r.logger.Error("Failed to scan message", "error", err)
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if channelID.Valid {
			message.ChannelID = &channelID.String
		}
		if replyToID.Valid {
			message.ReplyToID = &replyToID.String
		}
		if editedAt.Valid {
			message.EditedAt = &editedAt.Time
		}
		if deletedAt.Valid {
			message.DeletedAt = &deletedAt.Time
		}

		message.Sender = sender
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate messages: %w", err)
	}

	return messages, nil
}
