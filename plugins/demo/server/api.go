package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"

	"tgp/core/http"
	"tgp/core/i18n"
)

// handleTest обрабатывает тестовый GET эндпоинт.
func (s *Server) handleTest(w http.ResponseWriter, r *http.Request) {

	slog.Debug("test endpoint request",
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method),
	)

	s.writeJSON(w, http.StatusOK, map[string]string{
		"message": i18n.Msg("test endpoint works"),
		"status":  "ok",
	})
}

// handleEcho обрабатывает эхо POST эндпоинт.
func (s *Server) handleEcho(w http.ResponseWriter, r *http.Request) {

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

	data := make(map[string]interface{})
	for k, v := range values {
		if len(v) == 1 {
			data[k] = v[0]
		} else {
			data[k] = v
		}
	}

	s.writeJSON(w, http.StatusOK, EchoResult{
		Echo:   data,
		Status: "ok",
	})
}
