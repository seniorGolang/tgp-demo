package server

import (
	"tgp/core/http"
)

// handleStorage обрабатывает демонстрацию Storage.
func (s *Server) handleStorage(w http.ResponseWriter, r *http.Request) {

	s.writeHTML(w, http.StatusOK, formatResult(s.request))
}
