package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kseilons/messenger-backend/internal/kafka"
	"github.com/kseilons/messenger-backend/internal/models"
	"github.com/kseilons/messenger-backend/internal/service"
	ws "github.com/kseilons/messenger-backend/internal/websocket"
)

// CreateMessageRequest represents a request to create a message
type CreateMessageRequest struct {
	GroupID     string  `json:"group_id" binding:"required"`
	ChannelID   *string `json:"channel_id"`
	Content     string  `json:"content" binding:"required"`
	MessageType string  `json:"message_type"`
	ReplyToID   *string `json:"reply_to_id"`
}

// AddReactionRequest represents a request to add a reaction
type AddReactionRequest struct {
	Emoji string `json:"emoji" binding:"required"`
}

// RemoveReactionRequest represents a request to remove a reaction
type RemoveReactionRequest struct {
	Emoji string `json:"emoji" binding:"required"`
}

// CreateMessage creates a new message
func CreateMessage(messageService service.MessageService, wsHub *ws.Hub, kafkaProducer *kafka.Producer, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Invalid create message request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		serviceReq := &service.CreateMessageRequest{
			GroupID:     req.GroupID,
			ChannelID:   req.ChannelID,
			Content:     req.Content,
			MessageType: req.MessageType,
			ReplyToID:   req.ReplyToID,
		}

		message, err := messageService.CreateMessage(c.Request.Context(), serviceReq)
		if err != nil {
			logger.Error("Failed to create message", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
			return
		}

		// Broadcast message via WebSocket
		wsMessage := models.WebSocketMessage{
			Type:      models.WSMessageTypeNewMessage,
			Data:      message,
			Timestamp: time.Now(),
		}

		messageBytes, err := json.Marshal(wsMessage)
		if err != nil {
			logger.Error("Failed to marshal WebSocket message", "error", err)
		} else {
			// Broadcast to group room
			roomID := req.GroupID
			if req.ChannelID != nil {
				roomID = *req.ChannelID
			}
			wsHub.BroadcastToRoom(roomID, messageBytes)
		}

		// Publish to Kafka if enabled
		if kafkaProducer != nil {
			if err := kafkaProducer.PublishMessageEvent(models.KafkaEventTypeMessageCreated, message); err != nil {
				logger.Error("Failed to publish message event to Kafka", "error", err)
			}
		}

		logger.Info("Message created", "message_id", message.ID, "group_id", req.GroupID)
		c.JSON(http.StatusCreated, message)
	}
}

// GetMessagesByGroup retrieves messages for a group
func GetMessagesByGroup(messageService service.MessageService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID := c.Param("group_id")
		if groupID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
			return
		}

		// Parse query parameters
		limitStr := c.DefaultQuery("limit", "50")
		offsetStr := c.DefaultQuery("offset", "0")

		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
			return
		}

		messages, err := messageService.GetMessagesByGroup(c.Request.Context(), groupID, limit, offset)
		if err != nil {
			logger.Error("Failed to get messages by group", "error", err, "group_id", groupID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"messages": messages,
			"total":    len(messages),
			"limit":    limit,
			"offset":   offset,
		})
	}
}

// GetMessagesByChannel retrieves messages for a channel
func GetMessagesByChannel(messageService service.MessageService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID := c.Param("channel_id")
		if channelID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Channel ID is required"})
			return
		}

		// Parse query parameters
		limitStr := c.DefaultQuery("limit", "50")
		offsetStr := c.DefaultQuery("offset", "0")

		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
			return
		}

		messages, err := messageService.GetMessagesByChannel(c.Request.Context(), channelID, limit, offset)
		if err != nil {
			logger.Error("Failed to get messages by channel", "error", err, "channel_id", channelID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"messages": messages,
			"total":    len(messages),
			"limit":    limit,
			"offset":   offset,
		})
	}
}

// UpdateMessage updates a message
func UpdateMessage(messageService service.MessageService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageID := c.Param("id")
		if messageID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Message ID is required"})
			return
		}

		var req struct {
			Content string `json:"content" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Invalid update message request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Get user ID from JWT token
		userID := "temp-user-id"

		message, err := messageService.UpdateMessage(c.Request.Context(), messageID, req.Content, userID)
		if err != nil {
			logger.Error("Failed to update message", "error", err, "message_id", messageID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
			return
		}

		logger.Info("Message updated", "message_id", messageID, "user_id", userID)
		c.JSON(http.StatusOK, message)
	}
}

// DeleteMessage deletes a message
func DeleteMessage(messageService service.MessageService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageID := c.Param("id")
		if messageID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Message ID is required"})
			return
		}

		// TODO: Get user ID from JWT token
		userID := "temp-user-id"

		if err := messageService.DeleteMessage(c.Request.Context(), messageID, userID); err != nil {
			logger.Error("Failed to delete message", "error", err, "message_id", messageID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
			return
		}

		logger.Info("Message deleted", "message_id", messageID, "user_id", userID)
		c.JSON(http.StatusNoContent, nil)
	}
}

// AddReaction adds a reaction to a message
func AddReaction(messageService service.MessageService, wsHub *ws.Hub, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageID := c.Param("id")
		if messageID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Message ID is required"})
			return
		}

		var req AddReactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Invalid add reaction request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Get user ID from JWT token
		userID := "temp-user-id"

		reaction, err := messageService.AddReaction(c.Request.Context(), messageID, userID, req.Emoji)
		if err != nil {
			logger.Error("Failed to add reaction", "error", err, "message_id", messageID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reaction"})
			return
		}

		// Broadcast reaction via WebSocket
		wsMessage := models.WebSocketMessage{
			Type:      models.WSMessageTypeNewReaction,
			Data:      reaction,
			Timestamp: time.Now(),
		}

		messageBytes, err := json.Marshal(wsMessage)
		if err != nil {
			logger.Error("Failed to marshal WebSocket reaction message", "error", err)
		} else {
			// Get message to find room
			message, err := messageService.GetMessage(c.Request.Context(), messageID)
			if err == nil && message != nil {
				roomID := message.GroupID
				if message.ChannelID != nil {
					roomID = *message.ChannelID
				}
				wsHub.BroadcastToRoom(roomID, messageBytes)
			}
		}

		logger.Info("Reaction added", "message_id", messageID, "user_id", userID, "emoji", req.Emoji)
		c.JSON(http.StatusCreated, reaction)
	}
}

// RemoveReaction removes a reaction from a message
func RemoveReaction(messageService service.MessageService, wsHub *ws.Hub, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageID := c.Param("id")
		if messageID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Message ID is required"})
			return
		}

		var req RemoveReactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Invalid remove reaction request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Get user ID from JWT token
		userID := "temp-user-id"

		if err := messageService.RemoveReaction(c.Request.Context(), messageID, userID, req.Emoji); err != nil {
			logger.Error("Failed to remove reaction", "error", err, "message_id", messageID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove reaction"})
			return
		}

		// Broadcast reaction removal via WebSocket
		wsMessage := models.WebSocketMessage{
			Type: models.WSMessageTypeRemoveReaction,
			Data: map[string]interface{}{
				"message_id": messageID,
				"user_id":    userID,
				"emoji":      req.Emoji,
			},
			Timestamp: time.Now(),
		}

		messageBytes, err := json.Marshal(wsMessage)
		if err != nil {
			logger.Error("Failed to marshal WebSocket reaction removal message", "error", err)
		} else {
			// Get message to find room
			message, err := messageService.GetMessage(c.Request.Context(), messageID)
			if err == nil && message != nil {
				roomID := message.GroupID
				if message.ChannelID != nil {
					roomID = *message.ChannelID
				}
				wsHub.BroadcastToRoom(roomID, messageBytes)
			}
		}

		logger.Info("Reaction removed", "message_id", messageID, "user_id", userID, "emoji", req.Emoji)
		c.JSON(http.StatusNoContent, nil)
	}
}
