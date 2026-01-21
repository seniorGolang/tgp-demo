package server

import (
	"log/slog"

	"tgp/core/data"
	"tgp/core/http"
)

// NewServer создает новый сервер демо.
func NewServer(rootDir string, request data.Storage) (s *Server) {

	s = &Server{
		rootDir: rootDir,
		request: request,
		tasks:   make(map[uint32]*TaskState),
	}
	return s
}

// Start запускает HTTP сервер на указанном адресе.
func (s *Server) Start(addr string) (err error) {

	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/style.css", s.handleStyleCSS)
	mux.HandleFunc("/api/demo/storage", s.handleStorage)
	mux.HandleFunc("/api/demo/logging", s.handleLogging)
	mux.HandleFunc("/api/demo/filesystem", s.handleFilesystem)
	mux.HandleFunc("/api/demo/http-client", s.handleHTTPClient)
	mux.HandleFunc("/api/demo/server/status", s.handleServerStatus)
	mux.HandleFunc("/api/demo/command", s.handleCommand)
	mux.HandleFunc("/api/demo/plan", s.handlePlan)
	mux.HandleFunc("/api/demo/host-info", s.handleHostInfo)
	mux.HandleFunc("/api/demo/env", s.handleEnv)
	mux.HandleFunc("/api/demo/minimal", s.handleMinimal)
	mux.HandleFunc("/api/demo/interactive-select", s.handleInteractiveSelect)
	mux.HandleFunc("/api/demo/file-hash", s.handleFileHash)
	mux.HandleFunc("/api/demo/stop", s.handleStop)
	mux.HandleFunc("/api/demo/task/counter", s.handleTaskCounter)
	mux.HandleFunc("/api/demo/task/monitor", s.handleTaskMonitor)
	mux.HandleFunc("/api/demo/task/stats", s.handleTaskStats)
	mux.HandleFunc("/api/demo/tasks/status", s.handleTasksStatus)
	mux.HandleFunc("/api/demo/task/stop", s.handleTaskStop)
	mux.HandleFunc("/api/demo/tasks/stop-all", s.handleTasksStopAll)
	mux.HandleFunc("/api/test", s.handleTest)
	mux.HandleFunc("/api/echo", s.handleEcho)

	slog.Info("starting demo server", slog.String("addr", addr))

	var serverID uint64
	if serverID, err = http.ListenAndServe(addr, mux); err != nil {
		slog.Error("failed to start demo server", slog.String("addr", addr), slog.Any("error", err))
		return
	}

	s.serverID = serverID

	slog.Info("demo server started successfully", slog.String("addr", addr), slog.Uint64("serverID", serverID))
	return
}
