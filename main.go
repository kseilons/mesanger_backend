package main

import (
	"log/slog"

	"github.com/kseilons/messenger-backend/internal/config"
	"github.com/kseilons/messenger-backend/internal/logger"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Инициализация логгера
	log := logger.New(cfg.Log.ToLoggerConfig())

	// Создание сервера
	child := log.With(
		slog.Group("server", slog.Int("port", cfg.Server.Port)),
	)
	// Запуск сервера
	child.Info("Starting messenger server")

}
