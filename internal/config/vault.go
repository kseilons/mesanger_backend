package config

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

// VaultClient клиент для работы с Vault
type VaultClient struct {
	client    *vault.Client
	mountPath string
}

// NewVaultClient создает нового клиента Vault
func NewVaultClient(cfg *VaultConfig) (*VaultClient, error) {
	config := vault.DefaultConfig()
	config.Address = cfg.Address

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	client.SetToken(cfg.Token)

	if cfg.Namespace != "" {
		client.SetNamespace(cfg.Namespace)
	}

	return &VaultClient{
		client:    client,
		mountPath: cfg.MountPath,
	}, nil
}

// GetSecret получает секрет из Vault
func (vc *VaultClient) GetSecret(path string) (map[string]interface{}, error) {
	secret, err := vc.client.Logical().Read(vc.mountPath + "/data/" + path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret %s: %w", path, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret %s not found", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid secret format for %s", path)
	}

	return data, nil
}

// loadFromVault загружает секреты из Vault
func loadFromVault(cfg *Config) *Config {
	vaultClient, err := NewVaultClient(&cfg.Vault)
	if err != nil {
		log.Printf("Warning: Failed to initialize Vault client: %v", err)
		return cfg
	}

	v := reflect.ValueOf(cfg).Elem()
	loadSecretsFromVault(v, "", vaultClient)

	return cfg
}

// loadSecretsFromVault рекурсивно обходит структуру и загружает секреты из Vault
func loadSecretsFromVault(v reflect.Value, prefix string, vaultClient *VaultClient) {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		// Для вложенных структур рекурсивный вызов
		if field.Kind() == reflect.Struct {
			tag := fieldType.Tag.Get("vault")
			newPrefix := prefix
			if tag != "" {
				newPrefix = tag + "/"
			}
			loadSecretsFromVault(field, newPrefix, vaultClient)
			continue
		}

		// Получаем vault тег
		vaultPath := fieldType.Tag.Get("vault")
		if vaultPath == "" {
			continue
		}

		// Добавляем префикс если есть
		if prefix != "" {
			vaultPath = prefix + vaultPath
		}

		// Загружаем секрет из Vault
		secret, err := vaultClient.GetSecret(vaultPath)
		if err != nil {
			log.Printf("Warning: Failed to load secret %s: %v", vaultPath, err)
			continue
		}

		// Предполагаем, что ключ в секрете совпадает с именем поля
		fieldName := strings.ToLower(fieldType.Name)
		if value, exists := secret[fieldName]; exists {
			setFieldValue(field, value)
		}
	}
}

// setFieldValue устанавливает значение поля из Vault
func setFieldValue(field reflect.Value, value interface{}) {
	switch field.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		}
	case reflect.Int:
		if num, ok := value.(float64); ok {
			field.SetInt(int64(num))
		} else if str, ok := value.(string); ok {
			if intVal, err := strconv.Atoi(str); err == nil {
				field.SetInt(int64(intVal))
			}
		}
	case reflect.Bool:
		if b, ok := value.(bool); ok {
			field.SetBool(b)
		}
	default:
		panic("unhandled default case")
	}
}
