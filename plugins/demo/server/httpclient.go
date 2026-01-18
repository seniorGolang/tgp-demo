package server

import (
	"bytes"
	"fmt"
	"io"
	"net/url"

	"tgp/core/http"
	"tgp/core/i18n"
)

// handleHTTPClient обрабатывает демонстрацию HTTP клиента.
func (s *Server) handleHTTPClient(w http.ResponseWriter, r *http.Request) {

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

	method := values.Get("method")
	urlStr := values.Get("url")
	bodyStr := values.Get("body")

	if urlStr != "" && urlStr[0] == '/' {
		urlStr = "http://localhost:8080" + urlStr
	}

	var result HTTPClientResult

	var bodyReader io.Reader
	if bodyStr != "" {
		bodyReader = bytes.NewReader([]byte(bodyStr))
	}

	req, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to create request"), err)
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("HTTP request failed"), err)
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to read body"), err)
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	result.StatusCode = resp.StatusCode
	result.Headers = headers
	result.Body = string(bodyBytes)
	result.Message = i18n.Msg("HTTP request executed successfully")

	s.writeHTML(w, http.StatusOK, formatResult(result))
}
