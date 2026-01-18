package server

import (
	"tgp/core/http"
)

// handleIndex обрабатывает главную страницу.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {

	s.writeHTML(w, http.StatusOK, indexHTML)
}

// handleStyleCSS обрабатывает запрос CSS файла.
func (s *Server) handleStyleCSS(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(styleCSS))
}
