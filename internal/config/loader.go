package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

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
		Log: LogConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Vault: VaultConfig{
			Enabled:   false,
			MountPath: "secret",
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
		default:
			panic("unhandled default case")
		}
	}
}
