package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"

	"tgp/core"
	"tgp/core/data"
	"tgp/core/i18n"
	"tgp/core/plugin"
	"tgp/plugins/demo/server"
)

// DemoPlugin реализует интерфейс Plugin.
type DemoPlugin struct{}

// Execute выполняет основную логику плагина.
func (p *DemoPlugin) Execute(rootDir string, request data.Storage, path ...string) (response data.Storage, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in demo plugin: %v\n%s", r, debug.Stack())
			slog.Error("panic recovered in demo plugin",
				slog.Any("panic", r),
				slog.String("stack", string(debug.Stack())),
			)
			if response == nil {
				response = data.NewStorage()
			}
		}
	}()

	slog.Info(i18n.Msg("demo plugin started"))

	response = data.NewStorage()

	addr := ":8080"
	if request != nil {
		if addrRaw, ok := request.GetRaw("addr"); ok {
			var addrVal string
			if err = json.Unmarshal(addrRaw, &addrVal); err == nil && addrVal != "" {
				addr = addrVal
			}
		}
	}

	srv := server.NewServer(rootDir, request)

	defer func() {
		if cleanupErr := server.CleanupTempDir(); cleanupErr != nil {
			slog.Warn("failed to cleanup temp directory", slog.Any("error", cleanupErr))
		}
	}()

	if err = srv.Start(addr); err != nil {
		slog.Error("failed to start server", slog.Any("error", err))
		return
	}

	return
}

// Info возвращает информацию о плагине.
func (p *DemoPlugin) Info() (info plugin.Info, err error) {

	info = plugin.Info{
		Name:        "demo",
		Description: i18n.Msg("Demo plugin"),
		Author:      "AlexK <seniorGolang@gmail.com>",
		License:     "MIT",
		Category:    "utility",
		Commands: []plugin.Command{
			{
				Path:        []string{"demo", "serve"},
				Description: i18n.Msg("Start interactive demonstration of plugin capabilities"),
				Options: []plugin.Option{
					{
						Name:        "addr",
						Short:       "a",
						Type:        "string",
						Description: i18n.Msg("Address for HTTP server"),
						Required:    false,
						Default:     ":8080",
					},
				},
			},
		},
		AllowedHosts:     []string{"localhost", "127.0.0.1", "httpbin.org"},
		AllowedShellCMDs: []string{"uname", "go", "date"},
		AllowedEnvVars:   []string{"PATH", "HOME", "USER", "GOROOT", "GOPATH"},
		AllowedPaths: map[string]string{
			"@tg/tmp": "w",
		},
	}
	return
}

func init() {

	core.InitPlugin(&DemoPlugin{})
}

func main() {

	// Инициализация не требуется для wasip1
}
