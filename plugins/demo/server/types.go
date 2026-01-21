package server

import (
	"sync"
	"time"

	"tgp/core/data"
)

// executionPlan представляет план выполнения согласно документации.
type executionPlan struct {
	Current     int        `json:"current"`     // Индекс текущего шага в массиве Steps
	Steps       []planStep `json:"steps"`       // Результат планирования, включающий все шаги
	CommandPath []string   `json:"commandPath"` // Путь команды
	CommandArgs []string   `json:"commandArgs"` // Аргументы команды
}

// planStep представляет шаг плана выполнения согласно документации.
type planStep struct {
	Name         string        `json:"name"`                   // Имя плагина
	Kind         string        `json:"kind"`                   // Тип шага: "pre" | "stage" | "command" | "post"
	Version      string        `json:"version"`                // Версия плагина
	RequestKeys  []string      `json:"requestKeys,omitempty"`  // Ключи, которые были в request плагина (только для выполненных шагов)
	ResponseKeys []string      `json:"responseKeys,omitempty"` // Ключи, которые были в response плагина (только для выполненных шагов)
	Duration     time.Duration `json:"duration,omitempty"`     // Длительность выполнения плагина (только для выполненных шагов)
}

// Server управляет HTTP сервером демо.
type Server struct {
	rootDir  string
	request  data.Storage
	serverID uint64
	tasks    map[uint32]*TaskState
	tasksMu  sync.RWMutex
}

// TaskState представляет состояние задачи.
type TaskState struct {
	TaskID        uint32
	Description   string
	Interval      time.Duration
	Executions    int
	MaxExecutions int
	LastExecution time.Time
	Status        string // "active" | "stopped" | "auto-stopped"
	Data          map[string]any
	Logs          []string
}

// LoggingResult представляет результат логирования.
type LoggingResult struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Note      string `json:"note"`
}

// EchoResult представляет результат эхо запроса.
type EchoResult struct {
	Echo   map[string]interface{} `json:"echo"`
	Status string                 `json:"status"`
}

// ServerStatusResult представляет результат статуса сервера.
type ServerStatusResult struct {
	Status     string   `json:"status"`
	Message    string   `json:"message"`
	ListenerID uint64   `json:"serverID,omitempty"`
	Endpoints  []string `json:"endpoints"`
}

// StopResult представляет результат остановки сервера.
type StopResult struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// InteractiveSelectResult представляет результат интерактивного выбора.
type InteractiveSelectResult struct {
	Error         string   `json:"error,omitempty"`
	Options       []string `json:"options,omitempty"`
	Prompt        string   `json:"prompt,omitempty"`
	SelectedIndex int      `json:"selectedIndex,omitempty"`
	SelectedValue string   `json:"selectedValue,omitempty"`
}

// FilesystemResult представляет результат работы с файловой системой.
type FilesystemResult struct {
	Error            string `json:"error,omitempty"`
	DirectoryCreated string `json:"directoryCreated,omitempty"`
	FileCreated      string `json:"fileCreated,omitempty"`
	FileContent      string `json:"fileContent,omitempty"`
	FileRead         string `json:"fileRead,omitempty"`
	Message          string `json:"message,omitempty"`
}

// CommandResult представляет результат выполнения команды.
type CommandResult struct {
	Error    string   `json:"error,omitempty"`
	Command  string   `json:"command,omitempty"`
	Args     []string `json:"args,omitempty"`
	ExitCode int      `json:"exitCode,omitempty"`
	Stdout   string   `json:"stdout,omitempty"`
	Stderr   string   `json:"stderr,omitempty"`
	Message  string   `json:"message,omitempty"`
}

// PlanResult представляет результат визуализации плана.
type PlanResult struct {
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Plan    interface{} `json:"plan,omitempty"`
	SVG     string      `json:"svg,omitempty"`
}

// HTTPClientResult представляет результат HTTP клиента.
type HTTPClientResult struct {
	Error      string            `json:"error,omitempty"`
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
	Message    string            `json:"message,omitempty"`
}

// HostInfoResult представляет результат информации о хосте.
type HostInfoResult struct {
	HostInfo  map[string]string `json:"hostInfo"`
	Timestamp string            `json:"timestamp"`
	Message   string            `json:"message"`
}

// FileHashResult представляет результат вычисления хэша файла.
type FileHashResult struct {
	Error      string `json:"error,omitempty"`
	FileName   string `json:"fileName,omitempty"`
	FileSize   int64  `json:"fileSize,omitempty"`
	HashSHA256 string `json:"hashSHA256,omitempty"`
	Message    string `json:"message,omitempty"`
}

// TaskResult представляет результат работы задачи.
type TaskResult struct {
	TaskID        uint32         `json:"taskID,omitempty"`
	Description   string         `json:"description,omitempty"`
	Interval      string         `json:"interval,omitempty"`
	Executions    int            `json:"executions,omitempty"`
	MaxExecutions int            `json:"maxExecutions,omitempty"`
	LastExecution string         `json:"lastExecution,omitempty"`
	Status        string         `json:"status,omitempty"` // "active" | "stopped" | "auto-stopped"
	Data          map[string]any `json:"data,omitempty"`
	Logs          []string       `json:"logs,omitempty"`
	Error         string         `json:"error,omitempty"`
	Message       string         `json:"message,omitempty"`
}

// TasksStatusResult представляет результат статуса всех задач.
type TasksStatusResult struct {
	ActiveTasks []TaskResult `json:"activeTasks"`
	TotalActive int          `json:"totalActive"`
	Message     string       `json:"message,omitempty"`
	Error       string       `json:"error,omitempty"`
}
