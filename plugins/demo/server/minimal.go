package server

import (
	"fmt"
	"time"

	"tgp/core/http"
)

var requestCounter int64

// handleMinimal - минимальный handler для тестирования проблем с памятью.
func (s *Server) handleMinimal(w http.ResponseWriter, r *http.Request) {

	requestCounter++
	timestamp := time.Now().UnixNano()
	responseBody := fmt.Sprintf("OK-%d-%d", requestCounter, timestamp)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(responseBody))
}
