package main

import (
	"log/slog"
	"net/http"

	"github.com/kseilons/messenger-backend/internal/api"
	"github.com/kseilons/messenger-backend/internal/config"
	"github.com/kseilons/messenger-backend/internal/logger"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Инициализация логгера
	log := logger.New(cfg.Log.ToLoggerConfig())

	// Создание сервера

	// Настройка debug endpoints
	debugHandler := api.NewDebugHandler(log)

	// Регистрация debug handlers
	http.HandleFunc("/debug/loglevel", debugHandler.SetLogLevelHandler)
	http.HandleFunc("/debug/loglevel/current", debugHandler.GetLogLevelHandler)
	http.HandleFunc("/health", debugHandler.HealthCheckHandler)
	child := log.With(
		slog.Group("server", slog.Int("port", cfg.Server.Port)),
	)
	// Запуск сервера
	child.Info("Starting messenger server")

}
