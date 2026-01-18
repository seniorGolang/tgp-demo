package server

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"tgp/core/exec"
	"tgp/core/http"
	"tgp/core/i18n"
)

// handleHostInfo обрабатывает запрос информации о хосте.
func (s *Server) handleHostInfo(w http.ResponseWriter, r *http.Request) {

	var result HostInfoResult
	info := make(map[string]string)
	result.HostInfo = info

	cmd := exec.Command("uname", "-a")
	cmd = cmd.Dir(s.rootDir)
	if err := cmd.Start(); err != nil {
		slog.Warn("failed to start uname command", slog.Any("error", err))
	} else {
		var stdoutPipe io.ReadCloser
		var err error
		if stdoutPipe, err = cmd.StdoutPipe(); err != nil {
			slog.Warn("failed to create stdout pipe for uname", slog.Any("error", err))
		} else {
			defer stdoutPipe.Close()

			var stdoutBytes []byte
			var readErr error
			if stdoutBytes, readErr = io.ReadAll(stdoutPipe); readErr != nil {
				slog.Warn("failed to read stdout from uname", slog.Any("error", readErr), slog.Int("bytesRead", len(stdoutBytes)))
			} else {
				if waitErr := cmd.Wait(); waitErr != nil {
					slog.Warn("uname command failed", slog.Any("error", waitErr))
				} else {
					output := string(stdoutBytes)
					info["os"] = output
					slog.Debug("uname command succeeded", slog.String("output", output), slog.Int("outputLen", len(output)))
				}
			}
		}
	}

	cmd2 := exec.Command("go", "version")
	cmd2 = cmd2.Dir(s.rootDir)
	if err := cmd2.Start(); err != nil {
		slog.Warn("failed to start go version command", slog.Any("error", err))
	} else {
		var stdoutPipe io.ReadCloser
		var err error
		if stdoutPipe, err = cmd2.StdoutPipe(); err != nil {
			slog.Warn("failed to create stdout pipe for go version", slog.Any("error", err))
		} else {
			defer stdoutPipe.Close()

			var stdoutBytes []byte
			var readErr error
			if stdoutBytes, readErr = io.ReadAll(stdoutPipe); readErr != nil {
				slog.Warn("failed to read stdout from go version", slog.Any("error", readErr), slog.Int("bytesRead", len(stdoutBytes)))
			} else {
				if waitErr := cmd2.Wait(); waitErr != nil {
					slog.Warn("go version command failed", slog.Any("error", waitErr))
				} else {
					output := string(stdoutBytes)
					info["goVersion"] = output
					slog.Debug("go version command succeeded", slog.String("output", output), slog.Int("outputLen", len(output)))
				}
			}
		}
	}

	cmd3 := exec.Command("date")
	cmd3 = cmd3.Dir(s.rootDir)
	if err := cmd3.Start(); err != nil {
		slog.Warn("failed to start date command", slog.Any("error", err))
	} else {
		var stdoutPipe io.ReadCloser
		var err error
		if stdoutPipe, err = cmd3.StdoutPipe(); err != nil {
			slog.Warn("failed to create stdout pipe for date", slog.Any("error", err))
		} else {
			defer stdoutPipe.Close()

			var stdoutBytes []byte
			var readErr error
			if stdoutBytes, readErr = io.ReadAll(stdoutPipe); readErr != nil {
				slog.Warn("failed to read stdout from date", slog.Any("error", readErr), slog.Int("bytesRead", len(stdoutBytes)))
			} else {
				if waitErr := cmd3.Wait(); waitErr != nil {
					slog.Warn("date command failed", slog.Any("error", waitErr))
				} else {
					output := string(stdoutBytes)
					info["date"] = output
					slog.Debug("date command succeeded", slog.String("output", output), slog.Int("outputLen", len(output)))
				}
			}
		}
	}

	result.Timestamp = time.Now().Format(time.RFC3339)
	result.Message = i18n.Msg("Host information retrieved successfully")

	html := formatResult(result)
	s.writeHTML(w, http.StatusOK, html)
}

// handleEnv обрабатывает демонстрацию работы с переменными окружения.
func (s *Server) handleEnv(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	envNames := []string{"PATH", "HOME", "USER", "GOROOT", "GOPATH", "TG_LANG"}

	var response []byte
	response = append(response, i18n.Msg("Environment variables:")+"\n"...)

	for _, name := range envNames {
		value := os.Getenv(name)
		if value != "" {
			response = append(response, fmt.Sprintf("%s=%s\n", name, value)...)
		}
	}

	_, _ = w.Write(response)
}
