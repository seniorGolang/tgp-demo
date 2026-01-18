package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"strings"

	"tgp/core/http"
	"tgp/core/i18n"
)

const (
	defaultFileName = "uploaded_file"
	unnamedFileName = "unnamed"
)

// handleFileHash обрабатывает загрузку файла и вычисление его хэша.
func (s *Server) handleFileHash(w http.ResponseWriter, r *http.Request) {

	var result FileHashResult

	slog.Info("handleFileHash: получен запрос",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	if r.Body == nil {
		slog.Error("handleFileHash: Body is nil")
		result.Error = i18n.Msg("request body is nil")
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("handleFileHash: failed to read body", slog.Any("error", err))
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to read request body"), err)
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	if len(bodyData) == 0 {
		result.Error = i18n.Msg("request body is empty")
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	var jsonRequest struct {
		FileName string `json:"fileName"`
		FileData string `json:"fileData"`
	}

	if err := json.Unmarshal(bodyData, &jsonRequest); err == nil && jsonRequest.FileData != "" {
		fileData, decodeErr := base64.StdEncoding.DecodeString(jsonRequest.FileData)
		if decodeErr != nil {
			slog.Error("handleFileHash: failed to decode base64", slog.Any("error", decodeErr))
			result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to decode base64 file data"), decodeErr)
			s.writeHTML(w, http.StatusOK, formatResult(result))
			return
		}

		result.FileName = jsonRequest.FileName
		if result.FileName == "" {
			result.FileName = defaultFileName
		}
		result.FileSize = int64(len(fileData))

		hasher := sha256.New()
		_, err = hasher.Write(fileData)
		if err != nil {
			result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to compute hash"), err)
			s.writeHTML(w, http.StatusOK, formatResult(result))
			return
		}

		hashBytes := hasher.Sum(nil)
		result.HashSHA256 = fmt.Sprintf("%x", hashBytes)
		result.Message = fmt.Sprintf(i18n.Msg("File hash '%s' computed successfully"), result.FileName)

		slog.Info("handleFileHash: хэш вычислен (base64)",
			slog.String("fileName", result.FileName),
			slog.Int64("fileSize", result.FileSize),
			slog.String("hash", result.HashSHA256),
		)

		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	var contentType string
	for key, values := range r.Header {
		if len(values) > 0 {
			keyLower := strings.ToLower(key)
			if keyLower == "content-type" {
				contentType = values[0]
				break
			}
		}
	}

	if contentType != "" {
		boundary := getBoundary(contentType)
		if boundary != "" {
			reader := multipart.NewReader(bytes.NewReader(bodyData), boundary)

			part, err := reader.NextPart()
			if err == nil {
				fileName := part.FileName()
				if fileName == "" {
					fileName = part.FormName()
					if fileName == "" {
						fileName = unnamedFileName
					}
				}
				result.FileName = fileName

				fileData, err := io.ReadAll(part)
				if err == nil {
					defer part.Close()

					result.FileSize = int64(len(fileData))

					hasher := sha256.New()
					_, err = hasher.Write(fileData)
					if err == nil {
						hashBytes := hasher.Sum(nil)
						result.HashSHA256 = fmt.Sprintf("%x", hashBytes)
						result.Message = fmt.Sprintf(i18n.Msg("File hash '%s' computed successfully"), fileName)

						slog.Info("handleFileHash: хэш вычислен (multipart)",
							slog.String("fileName", fileName),
							slog.Int64("fileSize", result.FileSize),
							slog.String("hash", result.HashSHA256),
						)

						s.writeHTML(w, http.StatusOK, formatResult(result))
						return
					}
				}
			}
		}
	}

	result.Error = i18n.Msg("unable to parse file. Please use JSON format with base64 encoded fileData or multipart/form-data")
	s.writeHTML(w, http.StatusOK, formatResult(result))
}

// getBoundary извлекает boundary из Content-Type заголовка.
func getBoundary(contentType string) (boundary string) {

	const boundaryPrefix = "boundary="
	startIdx := -1
	for i := 0; i < len(contentType)-len(boundaryPrefix); i++ {
		if contentType[i:i+len(boundaryPrefix)] == boundaryPrefix {
			startIdx = i + len(boundaryPrefix)
			break
		}
	}

	if startIdx == -1 {
		return ""
	}

	boundary = contentType[startIdx:]

	if len(boundary) > 0 && boundary[0] == '"' {
		boundary = boundary[1:]
		if len(boundary) > 0 && boundary[len(boundary)-1] == '"' {
			boundary = boundary[:len(boundary)-1]
		}
	}

	boundary = trimSpaceAndSemicolon(boundary)

	return boundary
}

// trimSpaceAndSemicolon удаляет пробелы и точку с запятой в конце строки.
func trimSpaceAndSemicolon(s string) (result string) {

	result = s
	for len(result) > 0 && (result[len(result)-1] == ' ' || result[len(result)-1] == ';') {
		result = result[:len(result)-1]
	}
	return result
}
