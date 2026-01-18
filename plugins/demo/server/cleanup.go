package server

import (
	"log/slog"
	"os"
)

const tempDir = "/tg/tmp/demo"

// CleanupTempDir удаляет все содержимое временной папки плагина demo.
func CleanupTempDir() (err error) {

	if _, err = os.Stat(tempDir); os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		slog.Debug("temp directory not accessible for cleanup",
			slog.String("dir", tempDir),
			slog.Any("error", err))
		return nil
	}

	if err = os.RemoveAll(tempDir); err != nil {
		slog.Debug("failed to remove temp directory (may not exist)",
			slog.String("dir", tempDir),
			slog.Any("error", err))
		return nil
	}

	slog.Debug("temp directory cleanup attempted", slog.String("dir", tempDir))
	return nil
}
