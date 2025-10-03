package kafka

import (
	"log/slog"

	"github.com/kseilons/messenger-backend/internal/config"
	"github.com/kseilons/messenger-backend/internal/models"
)

// Producer represents a Kafka producer (stub implementation)
type Producer struct {
	logger *slog.Logger
	config config.KafkaConfig
}

// NewProducer creates a new Kafka producer (stub implementation)
func NewProducer(cfg config.KafkaConfig, logger *slog.Logger) (*Producer, error) {
	logger.Info("Kafka producer initialized (stub)", "brokers", cfg.Brokers)
	return &Producer{
		logger: logger,
		config: cfg,
	}, nil
}

// PublishMessage publishes a message to a Kafka topic (stub)
func (p *Producer) PublishMessage(topic string, event *models.KafkaEvent) error {
	p.logger.Debug("Message published (stub)", "topic", topic, "event_type", event.Type, "event_id", event.ID)
	return nil
}

// PublishMessageEvent publishes a message event (stub)
func (p *Producer) PublishMessageEvent(eventType models.KafkaEventType, message *models.Message) error {
	p.logger.Debug("Message event published (stub)", "event_type", eventType, "message_id", message.ID)
	return nil
}

// PublishUserEvent publishes a user event (stub)
func (p *Producer) PublishUserEvent(eventType models.KafkaEventType, userID string, data map[string]interface{}) error {
	p.logger.Debug("User event published (stub)", "event_type", eventType, "user_id", userID)
	return nil
}

// PublishGroupEvent publishes a group event (stub)
func (p *Producer) PublishGroupEvent(eventType models.KafkaEventType, groupID string, data map[string]interface{}) error {
	p.logger.Debug("Group event published (stub)", "event_type", eventType, "group_id", groupID)
	return nil
}

// PublishNotification publishes a notification event (stub)
func (p *Producer) PublishNotification(notification *models.Notification) error {
	p.logger.Debug("Notification published (stub)", "notification_id", notification.ID, "user_id", notification.UserID)
	return nil
}

// Close closes the producer (stub)
func (p *Producer) Close() {
	p.logger.Info("Kafka producer closed (stub)")
}
