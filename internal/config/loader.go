package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load загружает конфигурацию из YAML файла и environment variables
func Load() *Config {
	configPath := getConfigPath()

	// Загружаем из YAML
	cfg := loadFromYAML(configPath)

	// Переопределяем из environment variables
	cfg = overrideFromEnv(cfg)

	// Загружаем секреты из Vault если включено
	if cfg.Vault.Enabled {
		cfg = loadFromVault(cfg)
	}

	return cfg
}

// getConfigPath возвращает путь к конфигурационному файлу
func getConfigPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}

	// Поиск в стандартных местах
	possiblePaths := []string{
		"config.yaml",
		"config.yml",
		"config/config.yaml",
		"config/config.yml",
		"/etc/messenger/config.yaml",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Конфиг по умолчанию
	return "config.yaml"
}

// loadFromYAML загружает конфигурацию из YAML файла
func loadFromYAML(path string) *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host:         "localhost",
			Port:         8080,
			GRPCPort:     50051,
			ReadTimeout:  30,
			WriteTimeout: 30,
			IdleTimeout:  60,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Name:     "messenger_db",
			SSLMode:  "disable",
			MaxConns: 10,
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		JWT: JWTConfig{
			ExpirationHours:       24,
			RefreshExpirationDays: 7,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Vault: VaultConfig{
			Enabled:   false,
			MountPath: "secret",
		},
		Features: FeatureFlags{
			WebSocketEnabled:  true,
			RateLimitEnabled:  false,
			DebugEnabled:      false,
			KafkaEnabled:      false,
			FileUploadEnabled: false,
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     false,
			PingPeriod:      54,
			PongWait:        60,
			WriteWait:       10,
			MaxMessageSize:  1048576, // 1MB
		},
		Kafka: KafkaConfig{
			Brokers:         []string{"localhost:9092"},
			GroupID:         "messenger-backend",
			AutoOffsetReset: "latest",
		},
		FileStorage: FileStorageConfig{
			Type:         "local",
			LocalPath:    "./uploads",
			MaxFileSize:  10485760, // 10MB
			AllowedTypes: []string{"image/jpeg", "image/png", "image/gif", "application/pdf"},
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Warning: Could not read config file %s: %v\n", path, err)
		fmt.Println("Using default configuration")
		return cfg
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		panic(fmt.Sprintf("Failed to parse config file %s: %v", path, err))
	}

	return cfg
}

// overrideFromEnv переопределяет значения из environment variables
func overrideFromEnv(cfg *Config) *Config {
	v := reflect.ValueOf(cfg).Elem()
	overrideStruct(v, "")
	return cfg
}

// overrideStruct рекурсивно обходит структуру и применяет environment variables
func overrideStruct(v reflect.Value, prefix string) {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Пропускаем неэкспортируемые поля
		if !field.CanSet() {
			continue
		}

		// Для вложенных структур рекурсивный вызов
		if field.Kind() == reflect.Struct {
			tag := fieldType.Tag.Get("env")
			newPrefix := prefix
			if tag != "" {
				newPrefix = tag + "_"
			}
			overrideStruct(field, newPrefix)
			continue
		}

		// Получаем env тег
		envVar := fieldType.Tag.Get("env")
		if envVar == "" {
			continue
		}

		// Добавляем префикс если есть
		if prefix != "" {
			envVar = prefix + envVar
		}

		// Получаем значение из environment
		envValue := os.Getenv(envVar)
		if envValue == "" {
			continue
		}

		// Устанавливаем значение в зависимости от типа
		switch field.Kind() {
		case reflect.String:
			field.SetString(envValue)
		case reflect.Int:
			if intVal, err := strconv.Atoi(envValue); err == nil {
				field.SetInt(int64(intVal))
			}
		case reflect.Bool:
			if boolVal, err := strconv.ParseBool(envValue); err == nil {
				field.SetBool(boolVal)
			}
		case reflect.Int64:
			if intVal, err := strconv.ParseInt(envValue, 10, 64); err == nil {
				field.SetInt(intVal)
			}
		case reflect.Slice:
			// Для слайсов строк (например, KAFKA_BROKERS)
			if field.Type().Elem().Kind() == reflect.String {
				// Разделяем по запятой и создаем слайс
				if envValue != "" {
					values := []string{}
					for _, v := range strings.Split(envValue, ",") {
						if trimmed := strings.TrimSpace(v); trimmed != "" {
							values = append(values, trimmed)
						}
					}
					field.Set(reflect.ValueOf(values))
				}
			}
		default:
			fmt.Printf("Warning: Unhandled field type %v for field %s\n", field.Kind(), fieldType.Name)
			panic("unhandled default case")
		}
	}
}
