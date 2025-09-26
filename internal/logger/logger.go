package logger

import (
	"log/slog"
	"os"
)

// Config конфигурация логгера
type Config struct {
	Level     slog.Level
	Format    string // "json" или "text"
	Output    string // "stdout", "stderr" или путь к файлу
	File      string
	AddSource bool
}

// New создает новый логгер
func New(cfg Config) *slog.Logger {
	// Настройка вывода
	var output *os.File
	switch cfg.Output {
	case "stderr":
		output = os.Stderr
	case "file":
		// Для файла нужно отдельно обработать
		output = os.Stdout // временно
	default:
		output = os.Stdout
	}

	// Настройка формата
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return slog.New(handler)
}
