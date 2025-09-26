package logger

import (
	"log/slog"
	"os"
	"sync"
)

// Logger обертка вокруг slog с дополнительными возможностями
type Logger struct {
	*slog.Logger
	levelVar *slog.LevelVar
	mu       sync.RWMutex
}

// Config конфигурация логгера
type Config struct {
	Level     slog.Level
	Format    string // "json" или "text"
	Output    string // "stdout", "stderr" или путь к файлу
	File      string
	AddSource bool
}

// New создает новый логгер
func New(cfg Config) *Logger {
	levelVar := &slog.LevelVar{}
	levelVar.Set(cfg.Level)

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
		Level:     levelVar,
		AddSource: cfg.AddSource,
	}
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return &Logger{
		Logger:   slog.New(handler),
		levelVar: levelVar,
	}
}

// SetLevel динамически меняет уровень логирования
func (l *Logger) SetLevel(level slog.Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.levelVar.Set(level)
}

// GetLevel возвращает текущий уровень логирования
func (l *Logger) GetLevel() slog.Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.levelVar.Level()
}

// WithContext создает логгер с контекстом
func (l *Logger) WithContext(fields ...interface{}) *Logger {
	return &Logger{
		Logger:   l.Logger.With(fields...),
		levelVar: l.levelVar,
	}
}
