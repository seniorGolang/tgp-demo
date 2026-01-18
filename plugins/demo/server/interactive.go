package server

import (
	"fmt"
	"io"
	"net/url"

	"tgp/core"
	"tgp/core/http"
	"tgp/core/i18n"
)

// handleInteractiveSelect обрабатывает демонстрацию интерактивного выбора.
func (s *Server) handleInteractiveSelect(w http.ResponseWriter, r *http.Request) {

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

	prompt := values.Get("prompt")
	if prompt == "" {
		prompt = i18n.Msg("Select an option:")
	}

	var options []string
	if opts, ok := values["options"]; ok {
		options = opts
	}

	var result InteractiveSelectResult

	if len(options) == 0 {
		result.Error = i18n.Msg("Options not provided")
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	selected, err := core.InteractiveSelect(prompt, options, false, nil)
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("Interactive selection error"), err)
		result.Options = options
		result.Prompt = prompt
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	var selectedIndex = -1
	var selectedValue string
	if len(selected) > 0 {
		selectedValue = selected[0]
		for i, opt := range options {
			if opt == selectedValue {
				selectedIndex = i
				break
			}
		}
	}

	result.Options = options
	result.SelectedIndex = selectedIndex
	result.SelectedValue = selectedValue

	s.writeHTML(w, http.StatusOK, formatResult(result))
}
