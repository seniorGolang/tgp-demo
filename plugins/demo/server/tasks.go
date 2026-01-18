package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"sort"
	"time"

	"tgp/core"
	"tgp/core/http"
	"tgp/core/i18n"
)

const maxLogsPerTask = 20

// formatTaskResult форматирует результат задачи для отображения в HTML без JSON.
func formatTaskResult(result TaskResult) (html string) {

	if result.Error != "" {
		return fmt.Sprintf(`<div class="error">%s</div>`, result.Error)
	}

	html = fmt.Sprintf(`<div class="success">%s</div>`, result.Message)

	if result.TaskID > 0 {
		var statusColor string
		var statusText string
		switch result.Status {
		case "active":
			statusColor = "#27ae60"
			statusText = i18n.Msg("Active")
		case "auto-stopped":
			statusColor = "#f39c12"
			statusText = i18n.Msg("Auto-stopped")
		default:
			statusColor = "#95a5a6"
			statusText = i18n.Msg("Stopped")
		}

		html += fmt.Sprintf(`<div style="margin-top: 10px; padding: 10px; background: #f8f9fa; border-radius: 5px;">
			<p><strong>%s:</strong> %d</p>
			<p><strong>%s:</strong> %s</p>
			<p><strong>%s:</strong> <span style="color: %s; font-weight: bold;">%s</span></p>
		</div>`, i18n.Msg("Task ID"), result.TaskID, i18n.Msg("Description"), result.Description, i18n.Msg("Status"), statusColor, statusText)
	}

	return html
}

// taskHandlerAdapter адаптирует core.TaskHandler для использования с core.StartTask.
// Поскольку core.StartTask это алиас для wasm.StartTask, который ожидает wasm.TaskHandler,
// а мы используем core.TaskHandler, нужна адаптация типов.
func taskHandlerAdapter(handler core.TaskHandler) func() bool {

	return func() bool {
		return handler()
	}
}

// handleTaskCounter обрабатывает запуск задачи-счетчика.
func (s *Server) handleTaskCounter(w http.ResponseWriter, r *http.Request) {

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

	intervalStr := values.Get("interval")
	if intervalStr == "" {
		intervalStr = "2s"
	}

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		s.writeError(w, http.StatusOK, fmt.Sprintf("%s: %v", i18n.Msg("invalid interval"), err))
		return
	}

	maxExecutions := 5
	if maxStr := values.Get("maxExecutions"); maxStr != "" {
		if _, err := fmt.Sscanf(maxStr, "%d", &maxExecutions); err != nil {
			maxExecutions = 5
		}
	}

	var result TaskResult

	state := &TaskState{
		Description:   i18n.Msg("Counter task"),
		Interval:      interval,
		Executions:    0,
		MaxExecutions: maxExecutions,
		LastExecution: time.Now(),
		Status:        "active",
		Data:          make(map[string]any),
		Logs:          []string{},
	}
	state.Data["counter"] = 0

	handler := s.createCounterTask(state, maxExecutions)
	taskID, err := core.StartTask(interval, taskHandlerAdapter(handler))
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to start task"), err)
		s.writeHTML(w, http.StatusOK, formatTaskResult(result))
		return
	}

	state.TaskID = taskID
	s.tasksMu.Lock()
	s.tasks[taskID] = state
	s.tasksMu.Unlock()

	result.TaskID = taskID
	result.Description = i18n.Msg("Counter task")
	result.Interval = interval.String()
	result.MaxExecutions = maxExecutions
	result.Status = "active"
	result.Message = i18n.Msg("Counter task started successfully")

	slog.Info("counter task started", slog.Any("taskID", taskID), slog.String("interval", interval.String()))

	s.writeHTML(w, http.StatusOK, formatTaskResult(result))
}

// createCounterTask создает обработчик задачи-счетчика.
func (s *Server) createCounterTask(state *TaskState, maxExecutions int) core.TaskHandler {

	return func() (next bool) {

		s.tasksMu.Lock()
		defer s.tasksMu.Unlock()

		if state.Status != "active" {
			return false
		}

		state.Executions++
		state.LastExecution = time.Now()
		state.Data["counter"] = state.Executions
		logMsg := fmt.Sprintf("[%s] %s: %d/%d", time.Now().Format("15:04:05"), i18n.Msg("Counter"), state.Executions, maxExecutions)
		state.Logs = append(state.Logs, logMsg)
		if len(state.Logs) > maxLogsPerTask {
			state.Logs = state.Logs[len(state.Logs)-maxLogsPerTask:]
		}

		slog.Info("counter task executed", slog.Any("taskID", state.TaskID), slog.Int("execution", state.Executions), slog.Int("max", maxExecutions))

		if state.Executions >= maxExecutions {
			state.Status = "auto-stopped"
			slog.Info("counter task auto-stopped", slog.Any("taskID", state.TaskID), slog.Int("executions", state.Executions))
			return false
		}

		return true
	}
}

// handleTaskMonitor обрабатывает запуск задачи мониторинга времени.
func (s *Server) handleTaskMonitor(w http.ResponseWriter, r *http.Request) {

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

	intervalStr := values.Get("interval")
	if intervalStr == "" {
		intervalStr = "3s"
	}

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		s.writeError(w, http.StatusOK, fmt.Sprintf("%s: %v", i18n.Msg("invalid interval"), err))
		return
	}

	var result TaskResult

	state := &TaskState{
		Description:   i18n.Msg("Time monitor task"),
		Interval:      interval,
		Executions:    0,
		MaxExecutions: 0,
		LastExecution: time.Now(),
		Status:        "active",
		Data:          make(map[string]any),
		Logs:          []string{},
	}
	state.Data["timestamps"] = []string{}

	handler := s.createMonitorTask(state)
	taskID, err := core.StartTask(interval, taskHandlerAdapter(handler))
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to start task"), err)
		s.writeHTML(w, http.StatusOK, formatTaskResult(result))
		return
	}

	state.TaskID = taskID
	s.tasksMu.Lock()
	s.tasks[taskID] = state
	s.tasksMu.Unlock()

	result.TaskID = taskID
	result.Description = i18n.Msg("Time monitor task")
	result.Interval = interval.String()
	result.Status = "active"
	result.Message = i18n.Msg("Time monitor task started successfully")

	slog.Info("monitor task started", slog.Any("taskID", taskID), slog.String("interval", interval.String()))

	s.writeHTML(w, http.StatusOK, formatTaskResult(result))
}

// createMonitorTask создает обработчик задачи мониторинга времени.
func (s *Server) createMonitorTask(state *TaskState) core.TaskHandler {

	return func() (next bool) {

		s.tasksMu.Lock()
		defer s.tasksMu.Unlock()

		if state.Status != "active" {
			return false
		}

		state.Executions++
		state.LastExecution = time.Now()
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		logMsg := fmt.Sprintf("[%s] %s: %s", time.Now().Format("15:04:05"), i18n.Msg("Time"), timestamp)

		if timestamps, ok := state.Data["timestamps"].([]string); ok {
			timestamps = append(timestamps, timestamp)
			if len(timestamps) > 10 {
				timestamps = timestamps[len(timestamps)-10:]
			}
			state.Data["timestamps"] = timestamps
		} else {
			state.Data["timestamps"] = []string{timestamp}
		}

		state.Logs = append(state.Logs, logMsg)
		if len(state.Logs) > maxLogsPerTask {
			state.Logs = state.Logs[len(state.Logs)-maxLogsPerTask:]
		}

		slog.Info("monitor task executed", slog.Any("taskID", state.TaskID), slog.String("timestamp", timestamp))

		return true
	}
}

// handleTaskStats обрабатывает запуск задачи сбора статистики.
func (s *Server) handleTaskStats(w http.ResponseWriter, r *http.Request) {

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

	intervalStr := values.Get("interval")
	if intervalStr == "" {
		intervalStr = "5s"
	}

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		s.writeError(w, http.StatusOK, fmt.Sprintf("%s: %v", i18n.Msg("invalid interval"), err))
		return
	}

	var result TaskResult

	state := &TaskState{
		Description:   i18n.Msg("Statistics collection task"),
		Interval:      interval,
		Executions:    0,
		MaxExecutions: 0,
		LastExecution: time.Now(),
		Status:        "active",
		Data:          make(map[string]any),
		Logs:          []string{},
	}
	state.Data["requestCount"] = 0
	state.Data["history"] = []map[string]any{}

	handler := s.createStatsTask(state)
	taskID, err := core.StartTask(interval, taskHandlerAdapter(handler))
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to start task"), err)
		s.writeHTML(w, http.StatusOK, formatTaskResult(result))
		return
	}

	state.TaskID = taskID
	s.tasksMu.Lock()
	s.tasks[taskID] = state
	s.tasksMu.Unlock()

	result.TaskID = taskID
	result.Description = i18n.Msg("Statistics collection task")
	result.Interval = interval.String()
	result.Status = "active"
	result.Message = i18n.Msg("Statistics collection task started successfully")

	slog.Info("stats task started", slog.Any("taskID", taskID), slog.String("interval", interval.String()))

	s.writeHTML(w, http.StatusOK, formatTaskResult(result))
}

// createStatsTask создает обработчик задачи сбора статистики.
func (s *Server) createStatsTask(state *TaskState) core.TaskHandler {

	return func() (next bool) {

		s.tasksMu.Lock()
		defer s.tasksMu.Unlock()

		if state.Status != "active" {
			return false
		}

		state.Executions++
		state.LastExecution = time.Now()

		requestCount := state.Executions * 2
		state.Data["requestCount"] = requestCount

		statEntry := map[string]any{
			"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
			"requestCount": requestCount,
			"executions":   state.Executions,
		}

		if history, ok := state.Data["history"].([]map[string]any); ok {
			history = append(history, statEntry)
			if len(history) > 10 {
				history = history[len(history)-10:]
			}
			state.Data["history"] = history
		} else {
			state.Data["history"] = []map[string]any{statEntry}
		}

		logMsg := fmt.Sprintf("[%s] %s: %d requests", time.Now().Format("15:04:05"), i18n.Msg("Statistics"), requestCount)
		state.Logs = append(state.Logs, logMsg)
		if len(state.Logs) > maxLogsPerTask {
			state.Logs = state.Logs[len(state.Logs)-maxLogsPerTask:]
		}

		slog.Info("stats task executed", slog.Any("taskID", state.TaskID), slog.Int("requestCount", requestCount))

		return true
	}
}

// handleTasksStatus обрабатывает запрос статуса всех задач.
func (s *Server) handleTasksStatus(w http.ResponseWriter, r *http.Request) {

	s.tasksMu.RLock()
	defer s.tasksMu.RUnlock()

	tasks := make([]TaskResult, 0, len(s.tasks))
	for _, state := range s.tasks {
		intervalStr := state.Interval.String()
		if intervalStr == "" {
			intervalStr = "0s"
		}

		tasks = append(tasks, TaskResult{
			TaskID:        state.TaskID,
			Description:   state.Description,
			Interval:      intervalStr,
			Executions:    state.Executions,
			MaxExecutions: state.MaxExecutions,
			LastExecution: state.LastExecution.Format("15:04:05.000"),
			Status:        state.Status,
		})
	}

	// Сортируем задачи по taskID для стабильного порядка
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].TaskID < tasks[j].TaskID
	})

	activeCount := 0
	for _, task := range tasks {
		if task.Status == "active" {
			activeCount++
		}
	}

	result := TasksStatusResult{
		ActiveTasks: tasks,
		TotalActive: activeCount,
		Message:     fmt.Sprintf(i18n.Msg("Total tasks: %d, active: %d"), len(tasks), activeCount),
	}

	s.writeHTML(w, http.StatusOK, formatTasksStatusResult(result))
}

// handleTaskStop обрабатывает остановку задачи по ID.
func (s *Server) handleTaskStop(w http.ResponseWriter, r *http.Request) {

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

	taskIDStr := values.Get("taskID")
	if taskIDStr == "" {
		s.writeError(w, http.StatusOK, i18n.Msg("taskID parameter is required"))
		return
	}

	var taskID uint32
	if _, err := fmt.Sscanf(taskIDStr, "%d", &taskID); err != nil {
		s.writeError(w, http.StatusOK, fmt.Sprintf("%s: %v", i18n.Msg("invalid taskID"), err))
		return
	}

	var result TaskResult

	err = core.StopTask(taskID)
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to stop task"), err)
		s.writeHTML(w, http.StatusOK, formatTaskResult(result))
		return
	}

	s.tasksMu.Lock()
	if state, exists := s.tasks[taskID]; exists {
		state.Status = "stopped"
		result.TaskID = taskID
		result.Description = state.Description
		result.Status = "stopped"
		result.Message = i18n.Msg("Task stopped successfully")
	}
	s.tasksMu.Unlock()

	slog.Info("task stopped", slog.Any("taskID", taskID))

	s.writeHTML(w, http.StatusOK, formatTaskResult(result))
}

// handleTasksStopAll обрабатывает остановку всех задач.
func (s *Server) handleTasksStopAll(w http.ResponseWriter, r *http.Request) {

	var result TasksStatusResult

	err := core.StopAll()
	if err != nil {
		result.Error = fmt.Sprintf("%s: %v", i18n.Msg("failed to stop all tasks"), err)
		s.writeHTML(w, http.StatusOK, formatTasksStatusResult(TasksStatusResult{Error: result.Error}))
		return
	}

	s.tasksMu.Lock()
	stoppedCount := 0
	for _, state := range s.tasks {
		if state.Status == "active" {
			state.Status = "stopped"
			stoppedCount++
		}
	}
	s.tasksMu.Unlock()

	result.TotalActive = 0
	result.Message = fmt.Sprintf(i18n.Msg("All tasks stopped. Stopped %d task(s)"), stoppedCount)

	slog.Info("all tasks stopped", slog.Int("stopped", stoppedCount))

	s.writeHTML(w, http.StatusOK, formatTasksStatusResult(result))
}

// formatTasksStatusResult форматирует результат статуса задач для отображения.
func formatTasksStatusResult(result TasksStatusResult) (html string) {

	if result.Error != "" {
		return fmt.Sprintf(`<div class="error">%s</div>`, result.Error)
	}

	html = fmt.Sprintf(`<div style="margin-bottom: 15px;">
		<strong>%s</strong>: <span style="color: %s; font-weight: bold;">%d</span>
	</div>`, i18n.Msg("Active tasks"), func() string {
		if result.TotalActive > 0 {
			return "#27ae60"
		}
		return "#95a5a6"
	}(), result.TotalActive)

	if len(result.ActiveTasks) == 0 {
		html += fmt.Sprintf(`<div style="padding: 15px; background: #f8f9fa; border-radius: 5px; color: #6c757d; text-align: center;">
			<i>%s</i>
		</div>`, i18n.Msg("No tasks running"))
		return html
	}

	html += `<table style="width: 100%%; border-collapse: collapse; margin-top: 15px;">
		<thead>
			<tr style="background: #2c3e50; color: white;">
				<th style="padding: 10px; text-align: left; border: 1px solid #34495e;">ID</th>
				<th style="padding: 10px; text-align: left; border: 1px solid #34495e;">%s</th>
				<th style="padding: 10px; text-align: left; border: 1px solid #34495e;">%s</th>
				<th style="padding: 10px; text-align: left; border: 1px solid #34495e;">%s</th>
				<th style="padding: 10px; text-align: left; border: 1px solid #34495e;">%s</th>
				<th style="padding: 10px; text-align: left; border: 1px solid #34495e;">%s</th>
				<th style="padding: 10px; text-align: left; border: 1px solid #34495e;">%s</th>
			</tr>
		</thead>
		<tbody>`

	html = fmt.Sprintf(html, i18n.Msg("Description"), i18n.Msg("Interval"), i18n.Msg("Executions"), i18n.Msg("Last execution"), i18n.Msg("Status"), i18n.Msg("Action"))

	for _, task := range result.ActiveTasks {
		var statusColor string
		var statusText string
		switch task.Status {
		case "active":
			statusColor = "#27ae60"
			statusText = i18n.Msg("Active")
		case "auto-stopped":
			statusColor = "#f39c12"
			statusText = i18n.Msg("Auto-stopped")
		default:
			statusColor = "#95a5a6"
			statusText = i18n.Msg("Stopped")
		}

		executionsText := fmt.Sprintf("%d", task.Executions)
		if task.MaxExecutions > 0 {
			executionsText = fmt.Sprintf("%d/%d", task.Executions, task.MaxExecutions)
		}

		actionButton := ""
		if task.Status == "active" {
			actionButton = fmt.Sprintf(`<button hx-post="/api/demo/task/stop" hx-vals='{"taskID":"%d"}' hx-swap="none" hx-on::after-request="handleTaskStop(event, %d)" style="padding: 5px 10px; background: #e74c3c; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 12px;">%s</button>`, task.TaskID, task.TaskID, i18n.Msg("Stop"))
		}

		html += fmt.Sprintf(`<tr style="background: %s;">
			<td style="padding: 8px; border: 1px solid #dee2e6;">%d</td>
			<td style="padding: 8px; border: 1px solid #dee2e6;">%s</td>
			<td style="padding: 8px; border: 1px solid #dee2e6;">%s</td>
			<td style="padding: 8px; border: 1px solid #dee2e6;">%s</td>
			<td style="padding: 8px; border: 1px solid #dee2e6;">%s</td>
			<td style="padding: 8px; border: 1px solid #dee2e6;"><span style="color: %s; font-weight: bold;">%s</span></td>
			<td style="padding: 8px; border: 1px solid #dee2e6;">%s</td>
		</tr>`, func() string {
			if task.Status == "active" {
				return "#f8f9fa"
			}
			return "#ffffff"
		}(), task.TaskID, task.Description, task.Interval, executionsText, task.LastExecution, statusColor, statusText, actionButton)
	}

	html += `</tbody></table>`

	if result.Message != "" {
		html += fmt.Sprintf(`<div style="margin-top: 15px; padding: 10px; background: #e8f5e9; border-radius: 5px; color: #2e7d32;">
			%s
		</div>`, result.Message)
	}

	return html
}
