package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kseilons/messenger-backend/internal/config"
	"github.com/kseilons/messenger-backend/internal/models"
)

// Cache interface for caching operations
type Cache interface {
	// User operations
	SetUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, userID string) (*models.User, error)
	DeleteUser(ctx context.Context, userID string) error
	SetUserStatus(ctx context.Context, userID string, status models.UserStatus) error
	GetUserStatus(ctx context.Context, userID string) (models.UserStatus, error)
	SetOnlineUsers(ctx context.Context, userIDs []string) error
	GetOnlineUsers(ctx context.Context) ([]string, error)

	// Message operations
	SetMessage(ctx context.Context, message *models.Message) error
	GetMessage(ctx context.Context, messageID string) (*models.Message, error)
	DeleteMessage(ctx context.Context, messageID string) error
	SetMessageReactions(ctx context.Context, messageID string, reactions []*models.MessageReaction) error
	GetMessageReactions(ctx context.Context, messageID string) ([]*models.MessageReaction, error)

	// Group operations
	SetGroup(ctx context.Context, group *models.Group) error
	GetGroup(ctx context.Context, groupID string) (*models.Group, error)
	DeleteGroup(ctx context.Context, groupID string) error
	SetGroupMembers(ctx context.Context, groupID string, members []*models.GroupMember) error
	GetGroupMembers(ctx context.Context, groupID string) ([]*models.GroupMember, error)

	// WebSocket operations
	SetUserConnections(ctx context.Context, userID string, connectionIDs []string) error
	GetUserConnections(ctx context.Context, userID string) ([]string, error)
	AddUserConnection(ctx context.Context, userID, connectionID string) error
	RemoveUserConnection(ctx context.Context, userID, connectionID string) error

	// Typing status
	SetTypingStatus(ctx context.Context, status *models.TypingStatus) error
	GetTypingStatus(ctx context.Context, groupID string) ([]*models.TypingStatus, error)
	ClearTypingStatus(ctx context.Context, userID, groupID string) error

	// Generic operations
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
}

// redisCache implements Cache interface
type redisCache struct {
	client *redis.Client
	logger *slog.Logger
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(cfg config.RedisConfig, logger *slog.Logger) (Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis cache initialized", "host", cfg.Host, "port", cfg.Port, "db", cfg.DB)
	return &redisCache{
		client: rdb,
		logger: logger,
	}, nil
}

// SetUser caches a user
func (c *redisCache) SetUser(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf("user:%s", user.ID)
	return c.Set(ctx, key, user, 24*time.Hour)
}

// GetUser retrieves a user from cache
func (c *redisCache) GetUser(ctx context.Context, userID string) (*models.User, error) {
	key := fmt.Sprintf("user:%s", userID)
	var user models.User
	err := c.Get(ctx, key, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUser removes a user from cache
func (c *redisCache) DeleteUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:%s", userID)
	return c.Delete(ctx, key)
}

// SetUserStatus caches user status
func (c *redisCache) SetUserStatus(ctx context.Context, userID string, status models.UserStatus) error {
	key := fmt.Sprintf("user:%s:status", userID)
	return c.Set(ctx, key, status, 1*time.Hour)
}

// GetUserStatus retrieves user status from cache
func (c *redisCache) GetUserStatus(ctx context.Context, userID string) (models.UserStatus, error) {
	key := fmt.Sprintf("user:%s:status", userID)
	var status models.UserStatus
	err := c.Get(ctx, key, &status)
	return status, err
}

// SetOnlineUsers caches online user IDs
func (c *redisCache) SetOnlineUsers(ctx context.Context, userIDs []string) error {
	key := "users:online"
	return c.Set(ctx, key, userIDs, 5*time.Minute)
}

// GetOnlineUsers retrieves online user IDs from cache
func (c *redisCache) GetOnlineUsers(ctx context.Context) ([]string, error) {
	key := "users:online"
	var userIDs []string
	err := c.Get(ctx, key, &userIDs)
	return userIDs, err
}

// SetMessage caches a message
func (c *redisCache) SetMessage(ctx context.Context, message *models.Message) error {
	key := fmt.Sprintf("message:%s", message.ID)
	return c.Set(ctx, key, message, 1*time.Hour)
}

// GetMessage retrieves a message from cache
func (c *redisCache) GetMessage(ctx context.Context, messageID string) (*models.Message, error) {
	key := fmt.Sprintf("message:%s", messageID)
	var message models.Message
	err := c.Get(ctx, key, &message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// DeleteMessage removes a message from cache
func (c *redisCache) DeleteMessage(ctx context.Context, messageID string) error {
	key := fmt.Sprintf("message:%s", messageID)
	return c.Delete(ctx, key)
}

// SetMessageReactions caches message reactions
func (c *redisCache) SetMessageReactions(ctx context.Context, messageID string, reactions []*models.MessageReaction) error {
	key := fmt.Sprintf("message:%s:reactions", messageID)
	return c.Set(ctx, key, reactions, 30*time.Minute)
}

// GetMessageReactions retrieves message reactions from cache
func (c *redisCache) GetMessageReactions(ctx context.Context, messageID string) ([]*models.MessageReaction, error) {
	key := fmt.Sprintf("message:%s:reactions", messageID)
	var reactions []*models.MessageReaction
	err := c.Get(ctx, key, &reactions)
	return reactions, err
}

// SetGroup caches a group
func (c *redisCache) SetGroup(ctx context.Context, group *models.Group) error {
	key := fmt.Sprintf("group:%s", group.ID)
	return c.Set(ctx, key, group, 24*time.Hour)
}

// GetGroup retrieves a group from cache
func (c *redisCache) GetGroup(ctx context.Context, groupID string) (*models.Group, error) {
	key := fmt.Sprintf("group:%s", groupID)
	var group models.Group
	err := c.Get(ctx, key, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// DeleteGroup removes a group from cache
func (c *redisCache) DeleteGroup(ctx context.Context, groupID string) error {
	key := fmt.Sprintf("group:%s", groupID)
	return c.Delete(ctx, key)
}

// SetGroupMembers caches group members
func (c *redisCache) SetGroupMembers(ctx context.Context, groupID string, members []*models.GroupMember) error {
	key := fmt.Sprintf("group:%s:members", groupID)
	return c.Set(ctx, key, members, 1*time.Hour)
}

// GetGroupMembers retrieves group members from cache
func (c *redisCache) GetGroupMembers(ctx context.Context, groupID string) ([]*models.GroupMember, error) {
	key := fmt.Sprintf("group:%s:members", groupID)
	var members []*models.GroupMember
	err := c.Get(ctx, key, &members)
	return members, err
}

// SetUserConnections caches user WebSocket connections
func (c *redisCache) SetUserConnections(ctx context.Context, userID string, connectionIDs []string) error {
	key := fmt.Sprintf("user:%s:connections", userID)
	return c.Set(ctx, key, connectionIDs, 1*time.Hour)
}

// GetUserConnections retrieves user WebSocket connections from cache
func (c *redisCache) GetUserConnections(ctx context.Context, userID string) ([]string, error) {
	key := fmt.Sprintf("user:%s:connections", userID)
	var connectionIDs []string
	err := c.Get(ctx, key, &connectionIDs)
	return connectionIDs, err
}

// AddUserConnection adds a WebSocket connection to user
func (c *redisCache) AddUserConnection(ctx context.Context, userID, connectionID string) error {

	// Get existing connections
	connections, _ := c.GetUserConnections(ctx, userID)

	// Add new connection if not exists
	found := false
	for _, conn := range connections {
		if conn == connectionID {
			found = true
			break
		}
	}

	if !found {
		connections = append(connections, connectionID)
		return c.SetUserConnections(ctx, userID, connections)
	}

	return nil
}

// RemoveUserConnection removes a WebSocket connection from user
func (c *redisCache) RemoveUserConnection(ctx context.Context, userID, connectionID string) error {

	// Get existing connections
	connections, err := c.GetUserConnections(ctx, userID)
	if err != nil {
		return err
	}

	// Remove connection
	var newConnections []string
	for _, conn := range connections {
		if conn != connectionID {
			newConnections = append(newConnections, conn)
		}
	}

	return c.SetUserConnections(ctx, userID, newConnections)
}

// SetTypingStatus caches typing status
func (c *redisCache) SetTypingStatus(ctx context.Context, status *models.TypingStatus) error {
	key := fmt.Sprintf("typing:%s:%s", status.GroupID, status.UserID)
	return c.Set(ctx, key, status, 30*time.Second)
}

// GetTypingStatus retrieves typing statuses for a group
func (c *redisCache) GetTypingStatus(ctx context.Context, groupID string) ([]*models.TypingStatus, error) {
	pattern := fmt.Sprintf("typing:%s:*", groupID)
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var statuses []*models.TypingStatus
	for _, key := range keys {
		var status models.TypingStatus
		if err := c.Get(ctx, key, &status); err == nil {
			statuses = append(statuses, &status)
		}
	}

	return statuses, nil
}

// ClearTypingStatus clears typing status for a user in a group
func (c *redisCache) ClearTypingStatus(ctx context.Context, userID, groupID string) error {
	key := fmt.Sprintf("typing:%s:%s", groupID, userID)
	return c.Delete(ctx, key)
}

// Set sets a key-value pair with expiration
func (c *redisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value by key
func (c *redisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get key: %w", err)
	}

	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key
func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists checks if a key exists
func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	return result > 0, err
}

// Expire sets expiration for a key
func (c *redisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}
