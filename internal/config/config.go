package config

import (
	"log/slog"
	"strings"

	"github.com/kseilons/messenger-backend/internal/logger"
)

// Config основная структура конфигурации
type Config struct {
	Server   ServerConfig   `yaml:"server" json:"server"`
	Database DatabaseConfig `yaml:"database" json:"database"`
	Redis    RedisConfig    `yaml:"redis" json:"redis"`
	JWT      JWTConfig      `yaml:"jwt" json:"jwt"`
	Log      LogConfig      `yaml:"log" json:"log"`
	Vault    VaultConfig    `yaml:"vault" json:"vault"`
	Features FeatureFlags   `yaml:"features" json:"features"`
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
	WebSocketEnabled bool `yaml:"websocket_enabled" json:"websocket_enabled" env:"WEBSOCKET_ENABLED"`
	RateLimitEnabled bool `yaml:"rate_limit_enabled" json:"rate_limit_enabled" env:"RATE_LIMIT_ENABLED"`
	DebugEnabled     bool `yaml:"debug_enabled" json:"debug_enabled" env:"DEBUG_ENABLED"`
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
