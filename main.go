package main

import (
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
	logger.InitGlobal(cfg.Log.ToLoggerConfig())

	// Создание сервера
	srv := server.New(cfg, log)

	// Настройка debug endpoints
	debugHandler := api.NewDebugHandler(log)

	// Регистрация debug handlers
	http.HandleFunc("/debug/loglevel", debugHandler.SetLogLevelHandler)
	http.HandleFunc("/debug/loglevel/current", debugHandler.GetLogLevelHandler)
	http.HandleFunc("/health", debugHandler.HealthCheckHandler)

	// Запуск сервера
	log.Info("Starting messenger server", "port", cfg.Server.Port)
	if err := srv.Start(); err != nil {
		log.Error("Server failed to start", "error", err)
	}
}
