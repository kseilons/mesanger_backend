package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/kseilons/messenger-backend/internal/logger"
)

// DebugHandler обработчик debug endpoints
type DebugHandler struct {
	logger *logger.Logger
	mu     sync.RWMutex
}

// NewDebugHandler создает новый обработчик debug
func NewDebugHandler(l *logger.Logger) *DebugHandler {
	return &DebugHandler{logger: l}
}

// LogLevelRequest запрос на изменение уровня логирования
type LogLevelRequest struct {
	Level string `json:"level"`
}

// LogLevelResponse ответ с уровнем логирования
type LogLevelResponse struct {
	OldLevel string `json:"old_level"`
	NewLevel string `json:"new_level"`
	Message  string `json:"message"`
}

// SetLogLevelHandler меняет уровень логирования
func (h *DebugHandler) SetLogLevelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LogLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Валидация уровня
	var newLevel slog.Level
	switch req.Level {
	case "debug":
		newLevel = slog.LevelDebug
	case "info":
		newLevel = slog.LevelInfo
	case "warn":
		newLevel = slog.LevelWarn
	case "error":
		newLevel = slog.LevelError
	default:
		http.Error(w, `Invalid level. Use: "debug", "info", "warn", "error"`, http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	oldLevel := h.logger.GetLevel()
	h.logger.SetLevel(newLevel)
	h.mu.Unlock()

	response := LogLevelResponse{
		OldLevel: oldLevel.String(),
		NewLevel: newLevel.String(),
		Message:  "Log level updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLogLevelHandler возвращает текущий уровень логирования
func (h *DebugHandler) GetLogLevelHandler(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	currentLevel := h.logger.GetLevel()
	h.mu.RUnlock()

	response := map[string]string{
		"level": currentLevel.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheckHandler проверка здоровья сервиса
func (h *DebugHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":    "healthy",
		"log_level": h.logger.GetLevel().String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
