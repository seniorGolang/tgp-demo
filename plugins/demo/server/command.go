package server

import (
	"fmt"
	"io"
	"net/url"

	"tgp/core/exec"
	"tgp/core/http"
	"tgp/core/i18n"
)

// handleCommand обрабатывает демонстрацию выполнения команд.
func (s *Server) handleCommand(w http.ResponseWriter, r *http.Request) {

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

	command := values.Get("command")
	if command == "" {
		s.writeError(w, http.StatusOK, i18n.Msg("command parameter is required"))
		return
	}

	args := values["args"]

	var result CommandResult

	cmd := exec.Command(command, args...)
	cmd = cmd.Dir(s.rootDir)

	if err := cmd.Start(); err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to start command"), err)
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to create stdout pipe"), err)
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to create stderr pipe"), err)
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	stdoutBytes, err := io.ReadAll(stdoutPipe)
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to read stdout"), err)
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}
	defer stdoutPipe.Close()

	stderrBytes, err := io.ReadAll(stderrPipe)
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to read stderr"), err)
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}
	defer stderrPipe.Close()

	waitErr := cmd.Wait()

	exitCode := cmd.ExitCode()

	result.Command = command
	result.Args = args
	result.ExitCode = exitCode
	result.Stdout = string(stdoutBytes)
	result.Stderr = string(stderrBytes)

	if waitErr != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("command execution failed"), waitErr)
	} else {
		result.Message = i18n.Msg("Command executed successfully")
	}

	s.writeHTML(w, http.StatusOK, formatResult(result))
}
