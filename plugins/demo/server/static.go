package server

import (
	_ "embed"
)

//go:embed static/index.html
var indexHTML string

//go:embed static/style.css
var styleCSS string
