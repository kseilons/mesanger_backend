package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"

	"github.com/kseilons/messenger-backend/internal/api/handlers"
	"github.com/kseilons/messenger-backend/internal/config"
	"github.com/kseilons/messenger-backend/internal/kafka"
	"github.com/kseilons/messenger-backend/internal/logger"
	"github.com/kseilons/messenger-backend/internal/repository"
	"github.com/kseilons/messenger-backend/internal/service"
	ws "github.com/kseilons/messenger-backend/internal/websocket"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Инициализация логгера
	log := logger.New(cfg.Log.ToLoggerConfig())

	// Инициализация базы данных
	db, err := initDatabase(cfg, log)
	if err != nil {
		log.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db, log)
	messageRepo := repository.NewMessageRepository(db, log)
	// TODO: Добавить остальные репозитории

	// Инициализация WebSocket хаба
	wsHub := ws.NewHub(log)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск WebSocket хаба в отдельной горутине
	go wsHub.Run(ctx)

	// Инициализация сервисов
	userService := service.NewUserService(userRepo, log)
	messageService := service.NewMessageService(messageRepo, log)
	// TODO: Добавить остальные сервисы

	// Инициализация Kafka (если включен)
	var kafkaProducer *kafka.Producer
	if cfg.Features.KafkaEnabled {
		kafkaProducer, err = kafka.NewProducer(cfg.Kafka, log)
		if err != nil {
			log.Error("Failed to initialize Kafka producer", "error", err)
			os.Exit(1)
		}
		defer kafkaProducer.Close()
	}

	// Инициализация HTTP роутера
	router := initRouter(cfg, wsHub, userService, messageService, kafkaProducer, log)

	// Создание HTTP сервера
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Запуск сервера в отдельной горутине
	go func() {
		log.Info("Starting HTTP server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	cancel() // Останавливаем WebSocket хаб

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("Server exited")
}

// initDatabase инициализирует подключение к базе данных
func initDatabase(cfg *config.Config, log *slog.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(cfg.Database.MaxConns)
	db.SetMaxIdleConns(cfg.Database.MaxConns / 2)
	db.SetConnMaxLifetime(time.Hour)

	// Проверка соединения
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Info("Database connected successfully")
	return db, nil
}

// initRouter инициализирует HTTP роутер
func initRouter(cfg *config.Config, wsHub *ws.Hub, userService service.UserService,
	messageService service.MessageService, kafkaProducer *kafka.Producer, log *slog.Logger) *gin.Engine {

	// Настройка Gin
	if !cfg.Features.DebugEnabled {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// WebSocket endpoint
	if cfg.Features.WebSocketEnabled {
		router.GET("/ws", func(c *gin.Context) {
			handleWebSocket(c, wsHub, log)
		})
	}

	// API routes
	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/health", handlers.HealthCheck)

		// User routes
		users := api.Group("/users")
		{
			users.POST("/", handlers.CreateUser(userService, log))
			users.GET("/:id", handlers.GetUser(userService, log))
			users.PUT("/:id", handlers.UpdateUser(userService, log))
			users.DELETE("/:id", handlers.DeleteUser(userService, log))
			users.GET("/", handlers.SearchUsers(userService, log))
		}

		// Message routes
		messages := api.Group("/messages")
		{
			messages.POST("/", handlers.CreateMessage(messageService, wsHub, kafkaProducer, log))
			messages.GET("/group/:group_id", handlers.GetMessagesByGroup(messageService, log))
			messages.GET("/channel/:channel_id", handlers.GetMessagesByChannel(messageService, log))
			messages.PUT("/:id", handlers.UpdateMessage(messageService, log))
			messages.DELETE("/:id", handlers.DeleteMessage(messageService, log))
			messages.POST("/:id/reactions", handlers.AddReaction(messageService, wsHub, log))
			messages.DELETE("/:id/reactions", handlers.RemoveReaction(messageService, wsHub, log))
		}

		// TODO: Добавить остальные роуты для групп, каналов, уведомлений
	}

	return router
}

// handleWebSocket обрабатывает WebSocket соединения
func handleWebSocket(c *gin.Context, hub *ws.Hub, log *slog.Logger) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // В продакшене нужно добавить проверку origin
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("Failed to upgrade WebSocket connection", "error", err)
		return
	}

	// TODO: Добавить аутентификацию пользователя из JWT токена
	client := ws.NewClient(conn, hub, log)
	hub.RegisterClient(client)

	// Запуск горутин для чтения и записи
	go client.WritePump()
	go client.ReadPump()
}
