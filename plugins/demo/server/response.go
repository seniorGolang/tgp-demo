package server

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"tgp/core/http"
)

// writeHTML записывает HTML ответ.
func (s *Server) writeHTML(w http.ResponseWriter, statusCode int, html string) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	htmlBytes := []byte(html)
	written, err := w.Write(htmlBytes)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to write HTML response: statusCode=%d htmlLength=%d written=%d error=%v", statusCode, len(html), written, err))
	}
}

// writeJSON записывает JSON ответ.
func (s *Server) writeJSON(w http.ResponseWriter, statusCode int, data any) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal JSON", slog.Any("error", err))
		return
	}
	_, _ = w.Write(jsonData)
}

// writeError записывает ошибку в формате HTML.
func (s *Server) writeError(w http.ResponseWriter, statusCode int, message string) {

	html := `<div class="error">` + message + `</div>`
	s.writeHTML(w, statusCode, html)
}
