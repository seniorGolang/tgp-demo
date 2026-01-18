package server

import (
	"fmt"
	"os"
	"path/filepath"

	"tgp/core/http"
	"tgp/core/i18n"
)

const (
	demoDir      = "/tg/tmp/demo"
	testFileName = "test.txt"
)

// handleFilesystem обрабатывает демонстрацию файловой системы.
func (s *Server) handleFilesystem(w http.ResponseWriter, r *http.Request) {

	var result FilesystemResult

	if err := os.MkdirAll(demoDir, 0755); err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to create directory"), err)
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}
	result.DirectoryCreated = demoDir

	testFile := filepath.Join(demoDir, testFileName)
	content := i18n.Msg("Hello from WASM plugin!") + "\n" + i18n.Msg("This file was created by the demo plugin.")
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to create file"), err)
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	result.FileCreated = testFile
	result.FileContent = content

	readContent, err := os.ReadFile(testFile)
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to read file"), err)
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	result.FileRead = string(readContent)
	result.Message = i18n.Msg("Files successfully created and read")

	s.writeHTML(w, http.StatusOK, formatResult(result))
}
