package config

import (
	"log/slog"
	"strings"

	"github.com/kseilons/messenger-backend/internal/logger"
)

// Config основная структура конфигурации
type Config struct {
	Server      ServerConfig      `yaml:"server" json:"server"`
	Database    DatabaseConfig    `yaml:"database" json:"database"`
	Redis       RedisConfig       `yaml:"redis" json:"redis"`
	JWT         JWTConfig         `yaml:"jwt" json:"jwt"`
	Log         LogConfig         `yaml:"log" json:"log"`
	Vault       VaultConfig       `yaml:"vault" json:"vault"`
	Features    FeatureFlags      `yaml:"features" json:"features"`
	WebSocket   WebSocketConfig   `yaml:"websocket" json:"websocket"`
	Kafka       KafkaConfig       `yaml:"kafka" json:"kafka"`
	FileStorage FileStorageConfig `yaml:"file_storage" json:"file_storage"`
}

// ServerConfig конфигурация сервера
type ServerConfig struct {
	Host         string `yaml:"host" json:"host" env:"SERVER_HOST"`
	Port         int    `yaml:"port" json:"port" env:"SERVER_PORT"`
	GRPCPort     int    `yaml:"grpc_port" json:"grpc_port" env:"GRPC_PORT"`
	ReadTimeout  int    `yaml:"read_timeout" json:"read_timeout" env:"READ_TIMEOUT"`
	WriteTimeout int    `yaml:"write_timeout" json:"write_timeout" env:"WRITE_TIMEOUT"`
	IdleTimeout  int    `yaml:"idle_timeout" json:"idle_timeout" env:"IDLE_TIMEOUT"`
}

// DatabaseConfig конфигурация базы данных
type DatabaseConfig struct {
	Host     string `yaml:"host" json:"host" env:"DB_HOST"`
	Port     int    `yaml:"port" json:"port" env:"DB_PORT"`
	User     string `yaml:"user" json:"user" env:"DB_USER"`
	Password string `yaml:"password" json:"password" env:"DB_PASSWORD" vault:"database/password"`
	Name     string `yaml:"name" json:"name" env:"DB_NAME"`
	SSLMode  string `yaml:"ssl_mode" json:"ssl_mode" env:"DB_SSL_MODE"`
	MaxConns int    `yaml:"max_conns" json:"max_conns" env:"DB_MAX_CONNS"`
}

// RedisConfig конфигурация Redis
type RedisConfig struct {
	Host     string `yaml:"host" json:"host" env:"REDIS_HOST"`
	Port     int    `yaml:"port" json:"port" env:"REDIS_PORT"`
	Password string `yaml:"password" json:"password" env:"REDIS_PASSWORD" vault:"redis/password"`
	DB       int    `yaml:"db" json:"db" env:"REDIS_DB"`
}

// JWTConfig конфигурация JWT
type JWTConfig struct {
	Secret                string `yaml:"secret" json:"secret" env:"JWT_SECRET" vault:"jwt/secret"`
	ExpirationHours       int    `yaml:"expiration_hours" json:"expiration_hours" env:"JWT_EXPIRATION_HOURS"`
	RefreshExpirationDays int    `yaml:"refresh_expiration_days" json:"refresh_expiration_days" env:"JWT_REFRESH_EXPIRATION_DAYS"`
}

// LogConfig конфигурация логирования
type LogConfig struct {
	Level     string `yaml:"level" json:"level" env:"LOG_LEVEL"`
	Format    string `yaml:"format" json:"format" env:"LOG_FORMAT"`
	Output    string `yaml:"output" json:"output" env:"LOG_OUTPUT"`
	File      string `yaml:"file" json:"file" env:"LOG_FILE"`
	AddSource bool   `yaml:"add_source" json:"add_source" env:"LOG_ADD_SOURCE"`
}

// VaultConfig конфигурация HashiCorp Vault
type VaultConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled" env:"VAULT_ENABLED"`
	Address   string `yaml:"address" json:"address" env:"VAULT_ADDR"`
	Token     string `yaml:"token" json:"token" env:"VAULT_TOKEN"`
	MountPath string `yaml:"mount_path" json:"mount_path" env:"VAULT_MOUNT_PATH"`
	Namespace string `yaml:"namespace" json:"namespace" env:"VAULT_NAMESPACE"`
}

// FeatureFlags флаги функциональности
type FeatureFlags struct {
	WebSocketEnabled  bool `yaml:"websocket_enabled" json:"websocket_enabled" env:"WEBSOCKET_ENABLED"`
	RateLimitEnabled  bool `yaml:"rate_limit_enabled" json:"rate_limit_enabled" env:"RATE_LIMIT_ENABLED"`
	DebugEnabled      bool `yaml:"debug_enabled" json:"debug_enabled" env:"DEBUG_ENABLED"`
	KafkaEnabled      bool `yaml:"kafka_enabled" json:"kafka_enabled" env:"KAFKA_ENABLED"`
	FileUploadEnabled bool `yaml:"file_upload_enabled" json:"file_upload_enabled" env:"FILE_UPLOAD_ENABLED"`
}

// WebSocketConfig конфигурация WebSocket
type WebSocketConfig struct {
	ReadBufferSize  int   `yaml:"read_buffer_size" json:"read_buffer_size" env:"WS_READ_BUFFER_SIZE"`
	WriteBufferSize int   `yaml:"write_buffer_size" json:"write_buffer_size" env:"WS_WRITE_BUFFER_SIZE"`
	CheckOrigin     bool  `yaml:"check_origin" json:"check_origin" env:"WS_CHECK_ORIGIN"`
	PingPeriod      int   `yaml:"ping_period" json:"ping_period" env:"WS_PING_PERIOD"`
	PongWait        int   `yaml:"pong_wait" json:"pong_wait" env:"WS_PONG_WAIT"`
	WriteWait       int   `yaml:"write_wait" json:"write_wait" env:"WS_WRITE_WAIT"`
	MaxMessageSize  int64 `yaml:"max_message_size" json:"max_message_size" env:"WS_MAX_MESSAGE_SIZE"`
}

// KafkaConfig конфигурация Kafka
type KafkaConfig struct {
	Brokers          []string    `yaml:"brokers" json:"brokers" env:"KAFKA_BROKERS"`
	GroupID          string      `yaml:"group_id" json:"group_id" env:"KAFKA_GROUP_ID"`
	AutoOffsetReset  string      `yaml:"auto_offset_reset" json:"auto_offset_reset" env:"KAFKA_AUTO_OFFSET_RESET"`
	SecurityProtocol string      `yaml:"security_protocol" json:"security_protocol" env:"KAFKA_SECURITY_PROTOCOL"`
	SASLMechanism    string      `yaml:"sasl_mechanism" json:"sasl_mechanism" env:"KAFKA_SASL_MECHANISM"`
	SASLUsername     string      `yaml:"sasl_username" json:"sasl_username" env:"KAFKA_SASL_USERNAME"`
	SASLPassword     string      `yaml:"sasl_password" json:"sasl_password" env:"KAFKA_SASL_PASSWORD" vault:"kafka/password"`
	Topics           KafkaTopics `yaml:"topics" json:"topics"`
}

// KafkaTopics конфигурация топиков Kafka
type KafkaTopics struct {
	Messages      string `yaml:"messages" json:"messages" env:"KAFKA_TOPIC_MESSAGES"`
	Notifications string `yaml:"notifications" json:"notifications" env:"KAFKA_TOPIC_NOTIFICATIONS"`
	UserEvents    string `yaml:"user_events" json:"user_events" env:"KAFKA_TOPIC_USER_EVENTS"`
	GroupEvents   string `yaml:"group_events" json:"group_events" env:"KAFKA_TOPIC_GROUP_EVENTS"`
}

// FileStorageConfig конфигурация файлового хранилища
type FileStorageConfig struct {
	Type         string   `yaml:"type" json:"type" env:"FILE_STORAGE_TYPE"`
	LocalPath    string   `yaml:"local_path" json:"local_path" env:"FILE_STORAGE_LOCAL_PATH"`
	S3Bucket     string   `yaml:"s3_bucket" json:"s3_bucket" env:"FILE_STORAGE_S3_BUCKET"`
	S3Region     string   `yaml:"s3_region" json:"s3_region" env:"FILE_STORAGE_S3_REGION"`
	S3AccessKey  string   `yaml:"s3_access_key" json:"s3_access_key" env:"FILE_STORAGE_S3_ACCESS_KEY"`
	S3SecretKey  string   `yaml:"s3_secret_key" json:"s3_secret_key" env:"FILE_STORAGE_S3_SECRET_KEY" vault:"file_storage/s3_secret_key"`
	MaxFileSize  int64    `yaml:"max_file_size" json:"max_file_size" env:"FILE_STORAGE_MAX_FILE_SIZE"`
	AllowedTypes []string `yaml:"allowed_types" json:"allowed_types" env:"FILE_STORAGE_ALLOWED_TYPES"`
}

// ToLoggerConfig преобразует в конфиг логгера
func (lc *LogConfig) ToLoggerConfig() logger.Config {
	level := slog.LevelInfo
	switch strings.ToLower(lc.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	if lc.Format == "" {
		lc.Format = "json"
	}

	if lc.Output == "" {
		lc.Output = "stdout"
	}

	return logger.Config{
		Level:     level,
		Format:    lc.Format,
		Output:    lc.Output,
		File:      lc.File,
		AddSource: lc.AddSource,
	}
}
