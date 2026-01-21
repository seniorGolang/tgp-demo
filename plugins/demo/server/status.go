package server

import (
	"fmt"
	"log/slog"

	"tgp/core/http"
	"tgp/core/i18n"
)

// handleServerStatus обрабатывает запрос статуса сервера.
func (s *Server) handleServerStatus(w http.ResponseWriter, r *http.Request) {

	result := ServerStatusResult{
		Status:     "running",
		Message:    i18n.Msg("HTTP server is running and operational"),
		ListenerID: s.serverID,
		Endpoints: []string{
			"GET  /",
			"GET  /style.css",
			"POST /api/demo/storage",
			"POST /api/demo/logging",
			"POST /api/demo/filesystem",
			"POST /api/demo/http-client",
			"GET  /api/demo/server/status",
			"POST /api/demo/command",
			"POST /api/demo/plan",
			"GET  /api/demo/host-info",
			"GET  /api/demo/env",
			"POST /api/demo/interactive-select",
			"POST /api/demo/file-hash",
			"POST /api/demo/task/counter",
			"POST /api/demo/task/monitor",
			"POST /api/demo/task/stats",
			"GET  /api/demo/tasks/status",
			"POST /api/demo/task/stop",
			"POST /api/demo/tasks/stop-all",
			"POST /api/demo/stop",
			"GET  /api/test",
			"POST /api/echo",
		},
	}

	s.writeHTML(w, http.StatusOK, formatResult(result))
}

// handleStop обрабатывает запрос на остановку сервера.
func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {

	slog.Info(fmt.Sprintf("stop server request: path=%s serverID=%d", r.URL.Path, s.serverID))

	result := StopResult{
		Message: i18n.Msg("Server will be stopped."),
		Status:  "stopping",
	}
	s.writeHTML(w, http.StatusOK, formatResult(result))

	if s.serverID != 0 {
		if stopErr := http.StopServerByID(s.serverID); stopErr != nil {
			slog.Error("failed to stop server", slog.Uint64("serverID", s.serverID), slog.Any("error", stopErr))
		} else {
			slog.Info("server stopped successfully", slog.Uint64("serverID", s.serverID))
			s.serverID = 0
		}
	} else {
		slog.Info("serverID not set, server will stop automatically when Execute completes")
	}

	if cleanupErr := CleanupTempDir(); cleanupErr != nil {
		slog.Warn("failed to cleanup temp directory after stop", slog.Any("error", cleanupErr))
	}
}
