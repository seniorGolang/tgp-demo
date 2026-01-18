package server

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"tgp/core/i18n"
)

const maxJSONSize = 100 * 1024

// formatResult форматирует результат для отображения в HTML.
func formatResult(result any) (html string) {

	if selectResult, ok := result.(InteractiveSelectResult); ok {
		if len(selectResult.Options) > 0 && selectResult.SelectedIndex >= 0 {
			html = formatInteractiveSelectResult(selectResult)
			jsonData, err := json.Marshal(result)
			if err == nil && len(jsonData) <= maxJSONSize {
				html += fmt.Sprintf(`<details style="margin-top: 15px;"><summary style="cursor: pointer; color: #7f8c8d;">%s</summary><pre class="json">%s</pre></details>`, i18n.Msg("Show JSON"), string(jsonData))
			}
			return html
		}
	}

	if planResult, ok := result.(PlanResult); ok && planResult.SVG != "" {
		jsonData, err := json.Marshal(result)
		if err != nil {
			slog.Error("failed to marshal result to JSON",
				slog.Any("error", err),
			)
			return fmt.Sprintf(`<div class="error">%s: %v</div>`, i18n.Msg("Formatting error"), err)
		}
		if len(jsonData) > maxJSONSize {
			jsonData = jsonData[:maxJSONSize]
			jsonData = append(jsonData, []byte("... (truncated)")...)
		}
		html = fmt.Sprintf(`<div class="result">%s</div><pre class="json">%s</pre>`, planResult.SVG, string(jsonData))
		return html
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		slog.Error("failed to marshal result to JSON",
			slog.Any("error", err),
		)
		return fmt.Sprintf(`<div class="error">%s: %v</div>`, i18n.Msg("Formatting error"), err)
	}

	if len(jsonData) > maxJSONSize {
		jsonData = jsonData[:maxJSONSize]
		jsonData = append(jsonData, []byte("... (truncated)")...)
	}

	html = fmt.Sprintf(`<pre class="json">%s</pre>`, string(jsonData))

	return html
}

// formatInteractiveSelectResult форматирует результат интерактивного выбора.
func formatInteractiveSelectResult(result InteractiveSelectResult) (html string) {

	html = `<div class="interactive-select-result"><ul style="list-style: none; padding: 0; margin: 0;">`

	for i, opt := range result.Options {
		if i == result.SelectedIndex {
			html += fmt.Sprintf(`<li style="padding: 8px; margin: 5px 0; background: #d5f4e6; border-left: 4px solid #27ae60; border-radius: 4px;">
				<span style="color: #27ae60; font-weight: bold;">✓</span> <strong>%s</strong>
			</li>`, opt)
		} else {
			html += fmt.Sprintf(`<li style="padding: 8px; margin: 5px 0; background: #f8f9fa; border-left: 4px solid #dee2e6; border-radius: 4px;">
				<span style="color: #95a5a6;">○</span> %s
			</li>`, opt)
		}
	}

	html += `</ul></div>`

	return html
}
