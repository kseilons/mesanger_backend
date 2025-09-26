package logger

import (
	"log/slog"
	"os"
	"path/filepath"
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
	var err error

	switch cfg.Output {
	case "stderr":
		output = os.Stderr
	case "file":
		if cfg.File == "" {
			output = os.Stdout
		} else {
			// Создаем директорию для файла логов
			os.MkdirAll(filepath.Dir(cfg.File), 0755)

			// Открываем файл для записи
			output, err = os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				os.Stderr.WriteString("Failed to open log file: " + err.Error() + "\n")
				output = os.Stdout
			}
		}
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
