package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"time"

	"tgp/core/http"
	"tgp/core/i18n"
)

// handleLogging обрабатывает демонстрацию логирования.
func (s *Server) handleLogging(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, http.StatusOK, fmt.Sprintf("%s: %v", i18n.Msg("failed to read body"), err))
		return
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		s.writeError(w, http.StatusOK, fmt.Sprintf("%s: %v", i18n.Msg("failed to parse form"), err))
		return
	}

	level := values.Get("level")
	if level == "" {
		s.writeError(w, http.StatusOK, i18n.Msg("level parameter is required"))
		return
	}

	message := i18n.Msg("Logging demonstration at level") + " " + level
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	switch level {
	case "debug":
		slog.Debug(message, slog.String("timestamp", timestamp), slog.String("level", "debug"))
	case "info":
		slog.Info(message, slog.String("timestamp", timestamp), slog.String("level", "info"))
	case "warn":
		slog.Warn(message, slog.String("timestamp", timestamp), slog.String("level", "warn"))
	case "error":
		slog.Error(message, slog.String("timestamp", timestamp), slog.String("level", "error"))
	default:
		slog.Info(message, slog.String("level", level), slog.String("timestamp", timestamp))
	}

	result := LoggingResult{
		Level:     level,
		Message:   message,
		Timestamp: timestamp,
		Note:      i18n.Msg("Log sent to host console. Check the console where the plugin is running."),
	}

	s.writeHTML(w, http.StatusOK, formatResult(result))
}
