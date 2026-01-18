package server

import (
	"fmt"
	"math"
	"strings"
	"time"

	svg "github.com/ajstarks/svgo"

	"tgp/core/i18n"
)

// generatePlanSVG генерирует SVG визуализацию плана выполнения.
func generatePlanSVG(plan executionPlan) (result string) {

	if len(plan.Steps) == 0 {
		var buf strings.Builder
		canvas := svg.New(&buf)
		canvas.Start(400, 150)
		canvas.Def()
		canvas.LinearGradient("bgGrad", 0, 0, 0, 100, []svg.Offcolor{
			{Offset: 0, Color: "#f8f9fa", Opacity: 1},
			{Offset: 100, Color: "#e9ecef", Opacity: 1},
		})
		canvas.DefEnd()
		canvas.Roundrect(0, 0, 400, 150, 8, 8, "fill:url(#bgGrad);stroke:#dee2e6;stroke-width:2")
		canvas.Text(200, 75, i18n.Msg("Plan not found"), "text-anchor:middle;fill:#6c757d;font-size:16;font-family:Arial, sans-serif")
		canvas.End()
		return buf.String()
	}

	graph := buildGraphFromPlan(plan)
	result = renderGraph(graph)

	return result
}

// graphNode представляет узел графа.
type graphNode struct {
	ID           string
	Label        string
	Type         string
	X            float64
	Y            float64
	Width        float64
	Height       float64
	Level        int
	IsCurrent    bool
	IsParallel   bool
	Version      string
	RequestKeys  []string
	ResponseKeys []string
	Duration     string
}

// graphEdge представляет связь между узлами.
type graphEdge struct {
	From  string
	To    string
	Type  string
	Label string
}

// graph представляет граф выполнения.
type graph struct {
	Nodes       []*graphNode
	Edges       []*graphEdge
	CommandPath []string
	CommandArgs []string
}

// buildGraphFromPlan строит граф из структуры плана.
func buildGraphFromPlan(plan executionPlan) (g *graph) {

	g = &graph{
		Nodes:       []*graphNode{},
		Edges:       []*graphEdge{},
		CommandPath: plan.CommandPath,
		CommandArgs: plan.CommandArgs,
	}

	extractNodesFromTypedPlan(plan, g)

	addStartNode(g)

	addStopNode(g)

	layoutNodes(g)

	return g
}

// extractNodesFromTypedPlan извлекает узлы из типизированного плана.
func extractNodesFromTypedPlan(plan executionPlan, g *graph) {

	for i, step := range plan.Steps {
		node := extractNodeFromPlanStep(step, i)
		if node != nil {
			if i == plan.Current {
				node.IsCurrent = true
			}
			g.Nodes = append(g.Nodes, node)
		}
	}

	for i := 0; i < len(g.Nodes)-1; i++ {
		g.Edges = append(g.Edges, &graphEdge{
			From: g.Nodes[i].ID,
			To:   g.Nodes[i+1].ID,
			Type: "sequential",
		})
	}
}

// extractNodeFromPlanStep извлекает узел из типизированного шага плана.
func extractNodeFromPlanStep(step planStep, index int) (node *graphNode) {

	node = &graphNode{
		ID:           fmt.Sprintf("node_%d", index),
		Label:        step.Name,
		Type:         normalizeKindToType(step.Kind),
		Level:        0,
		Version:      step.Version,
		RequestKeys:  step.RequestKeys,
		ResponseKeys: step.ResponseKeys,
		Duration:     formatDuration(step.Duration),
	}

	if node.Label == "" {
		node.Label = fmt.Sprintf(i18n.Msg("Step %d"), index+1)
	}

	return node
}

// normalizeKindToType преобразует kind плагина в тип для визуализации.
func normalizeKindToType(kind string) (normalized string) {

	kind = strings.ToLower(strings.TrimSpace(kind))

	typeMap := map[string]string{
		"pre":     "wasm",
		"stage":   "wasm",
		"command": "command",
		"post":    "wasm",
	}

	if normalized, ok := typeMap[kind]; ok {
		return normalized
	}

	return kind
}

// formatDuration форматирует длительность для отображения.
func formatDuration(d time.Duration) (formatted string) {

	if d == 0 {
		return ""
	}

	if d < time.Millisecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}

	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// addStartNode добавляет узел старта в начало графа.
func addStartNode(g *graph) {

	if len(g.Nodes) == 0 {
		return
	}

	startNode := &graphNode{
		ID:         "start",
		Label:      i18n.Msg("Start"),
		Type:       "start",
		X:          0,
		Y:          0,
		Width:      60,
		Height:     60,
		Level:      0,
		IsCurrent:  false,
		IsParallel: false,
	}

	g.Nodes = append([]*graphNode{startNode}, g.Nodes...)

	if len(g.Nodes) > 1 {
		g.Edges = append([]*graphEdge{
			{
				From:  "start",
				To:    g.Nodes[1].ID,
				Type:  "sequential",
				Label: "",
			},
		}, g.Edges...)
	}
}

// addStopNode добавляет узел стопа в конец графа.
func addStopNode(g *graph) {

	if len(g.Nodes) == 0 {
		return
	}

	var lastRealNode *graphNode
	for i := len(g.Nodes) - 1; i >= 0; i-- {
		if g.Nodes[i].Type != "start" && g.Nodes[i].Type != "stop" {
			lastRealNode = g.Nodes[i]
			break
		}
	}

	if lastRealNode == nil {
		return
	}

	stopNode := &graphNode{
		ID:         "stop",
		Label:      i18n.Msg("Stop"),
		Type:       "stop",
		X:          0,
		Y:          0,
		Width:      60,
		Height:     60,
		Level:      0,
		IsCurrent:  false,
		IsParallel: false,
	}

	g.Nodes = append(g.Nodes, stopNode)

	g.Edges = append(g.Edges, &graphEdge{
		From:  lastRealNode.ID,
		To:    "stop",
		Type:  "sequential",
		Label: "",
	})
}

// layoutNodes позиционирует узлы графа.
func layoutNodes(g *graph) {

	if len(g.Nodes) == 0 {
		return
	}

	nodeWidth := 150.0
	nodeHeight := 80.0
	spacingX := 200.0
	spacingY := 120.0
	startX := 100.0
	startY := 100.0

	var startNode *graphNode
	var stopNode *graphNode
	var regularNodes []*graphNode

	if len(g.Nodes) > 0 && g.Nodes[0].Type == "start" {
		startNode = g.Nodes[0]
		regularNodes = g.Nodes[1:]
	} else {
		regularNodes = g.Nodes
	}

	if len(regularNodes) > 0 && regularNodes[len(regularNodes)-1].Type == "stop" {
		stopNode = regularNodes[len(regularNodes)-1]
		regularNodes = regularNodes[:len(regularNodes)-1]
	}

	levels := make(map[int][]*graphNode)
	for _, node := range regularNodes {
		levels[node.Level] = append(levels[node.Level], node)
	}

	maxLevel := 0
	for level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	currentY := startY
	for level := 0; level <= maxLevel; level++ {
		nodesInLevel := levels[level]
		if len(nodesInLevel) == 0 {
			continue
		}

		totalWidth := float64(len(nodesInLevel)-1) * spacingX
		currentX := startX + (800.0-totalWidth-startX*2)/2

		for _, node := range nodesInLevel {
			node.X = currentX
			node.Y = currentY
			node.Width = nodeWidth
			node.Height = nodeHeight
			currentX += spacingX
		}

		currentY += spacingY
	}

	if maxLevel == 0 {
		gapBetweenElements := 50.0

		currentX := startX

		if startNode != nil {
			startNode.X = currentX
			startNode.Y = startY + (nodeHeight-60.0)/2
			startNode.Width = 60.0
			startNode.Height = 60.0
			currentX += startNode.Width + gapBetweenElements
		}

		for i := 0; i < len(regularNodes); i++ {
			node := regularNodes[i]
			node.X = currentX
			node.Y = startY
			node.Width = nodeWidth
			node.Height = nodeHeight
			currentX += node.Width + gapBetweenElements
		}

		if stopNode != nil {
			stopNode.X = currentX
			stopNode.Y = startY + (nodeHeight-60.0)/2
			stopNode.Width = 60.0
			stopNode.Height = 60.0
		}
	}
}

// renderGraph отрисовывает граф в SVG.
func renderGraph(g *graph) (result string) {

	if len(g.Nodes) == 0 {
		var buf strings.Builder
		canvas := svg.New(&buf)
		canvas.Start(400, 150)
		canvas.Roundrect(0, 0, 400, 150, 8, 8, "fill:#f8f9fa")
		canvas.Text(200, 75, i18n.Msg("No data to visualize"), "text-anchor:middle;fill:#999;font-size:14")
		canvas.End()
		return buf.String()
	}

	minX, minY, maxX, maxY := calculateBounds(g)
	padding := 50.0
	headerHeight := 30.0
	width := maxX - minX + padding*2
	height := maxY - minY + padding*2 + headerHeight

	offsetX := -minX + padding
	offsetY := -minY + padding + headerHeight

	var buf strings.Builder
	canvas := svg.New(&buf)
	canvas.Start(int(width), int(height))

	canvas.Def()

	fmt.Fprintf(&buf, `<marker id="arrowhead" markerWidth="8" markerHeight="8" refX="7" refY="2.5" orient="auto" viewBox="0 0 8 8">`)
	fmt.Fprintf(&buf, `<polygon points="0,0 8,2.5 0,5" fill="#90caf9" stroke="#90caf9" stroke-width="0.5"/>`)
	fmt.Fprintf(&buf, `</marker>`)

	canvas.LinearGradient("wasmGrad", 0, 0, 0, 100, []svg.Offcolor{
		{Offset: 0, Color: "#f3f7fc", Opacity: 1},
		{Offset: 100, Color: "#e8f0f8", Opacity: 1},
	})

	canvas.LinearGradient("nativeGrad", 0, 0, 0, 100, []svg.Offcolor{
		{Offset: 0, Color: "#f4f8f5", Opacity: 1},
		{Offset: 100, Color: "#eaf2eb", Opacity: 1},
	})

	canvas.LinearGradient("commandGrad", 0, 0, 0, 100, []svg.Offcolor{
		{Offset: 0, Color: "#fef8f3", Opacity: 1},
		{Offset: 100, Color: "#fdf1e6", Opacity: 1},
	})

	canvas.LinearGradient("taskGrad", 0, 0, 0, 100, []svg.Offcolor{
		{Offset: 0, Color: "#f9f5fa", Opacity: 1},
		{Offset: 100, Color: "#f3ebf5", Opacity: 1},
	})

	canvas.LinearGradient("transformGrad", 0, 0, 0, 100, []svg.Offcolor{
		{Offset: 0, Color: "#fdf5f8", Opacity: 1},
		{Offset: 100, Color: "#faebf0", Opacity: 1},
	})

	canvas.LinearGradient("defaultGrad", 0, 0, 0, 100, []svg.Offcolor{
		{Offset: 0, Color: "#fafafa", Opacity: 1},
		{Offset: 100, Color: "#f5f5f5", Opacity: 1},
	})

	canvas.LinearGradient("currentNodeGrad", 0, 0, 0, 100, []svg.Offcolor{
		{Offset: 0, Color: "#d5f4e6", Opacity: 1},
		{Offset: 100, Color: "#a8e6cf", Opacity: 1},
	})

	canvas.DefEnd()

	canvas.Roundrect(0, 0, int(width), int(height), 8, 8, "fill:#ffffff;stroke:#dee2e6;stroke-width:2")

	if len(g.CommandPath) > 0 || len(g.CommandArgs) > 0 {
		commandText := strings.Join(g.CommandPath, " ")
		if len(g.CommandArgs) > 0 {
			commandText += " " + strings.Join(g.CommandArgs, " ")
		}
		if len(commandText) > 80 {
			commandText = commandText[:77] + "..."
		}
		canvas.Text(int(width/2), 20, commandText, "text-anchor:middle;fill:#7f8c8d;font-size:10;font-family:monospace;font-style:italic")
	}

	for _, node := range g.Nodes {
		x := node.X + offsetX
		y := node.Y + offsetY

		if node.Type == "start" {
			centerX := int(x + node.Width/2)
			centerY := int(y + node.Height/2)
			radius := int(node.Width / 2)

			canvas.Circle(centerX+1, centerY+1, radius, "fill:rgba(0,0,0,0.1)")
			canvas.Circle(centerX, centerY, radius, "fill:#c8e6c9;stroke:#a5d6a7;stroke-width:1.5")
			canvas.Text(centerX, centerY+4, i18n.Msg("Start"), "text-anchor:middle;fill:#2e7d32;font-size:10;font-weight:bold;font-family:Arial, sans-serif")
			continue
		}

		if node.Type == "stop" {
			centerX := int(x + node.Width/2)
			centerY := int(y + node.Height/2)
			radius := int(node.Width / 2)

			canvas.Circle(centerX+1, centerY+1, radius, "fill:rgba(0,0,0,0.1)")
			canvas.Circle(centerX, centerY, radius, "fill:#fce4ec;stroke:#f8bbd0;stroke-width:1.5")
			canvas.Text(centerX, centerY+4, i18n.Msg("Stop"), "text-anchor:middle;fill:#c2185b;font-size:10;font-weight:bold;font-family:Arial, sans-serif")
			continue
		}

		gradientID, strokeColor := getNodeStyle(node)

		canvas.Roundrect(int(x+2), int(y+2), int(node.Width), int(node.Height), 8, 8, "fill:rgba(0,0,0,0.1)")

		strokeWidth := 1.0
		if node.IsCurrent {
			strokeWidth = 1.5
		}
		canvas.Roundrect(int(x), int(y), int(node.Width), int(node.Height), 8, 8, fmt.Sprintf("fill:url(#%s);stroke:%s;stroke-width:%.1f", gradientID, strokeColor, strokeWidth))

		label := node.Label
		labelY := y + node.Height/2

		if node.Version != "" {
			labelY = y + node.Height/2 - 10
		}

		if len(label) > 20 {
			words := strings.Fields(label)
			if len(words) > 0 {
				lines := []string{}
				currentLine := ""
				for _, word := range words {
					if len(currentLine)+len(word)+1 <= 20 {
						if currentLine != "" {
							currentLine += " "
						}
						currentLine += word
					} else {
						if currentLine != "" {
							lines = append(lines, currentLine)
						}
						currentLine = word
					}
				}
				if currentLine != "" {
					lines = append(lines, currentLine)
				}
				if len(lines) > 2 {
					lines = lines[:2]
				}
				textStartY := labelY - float64(len(lines)-1)*9
				for i, line := range lines {
					canvas.Text(int(x+node.Width/2), int(textStartY+float64(i)*18), line, "text-anchor:middle;fill:#2c3e50;font-size:12;font-weight:bold;font-family:Arial, sans-serif")
				}
			} else {
				canvas.Text(int(x+node.Width/2), int(labelY+4), label, "text-anchor:middle;fill:#2c3e50;font-size:12;font-weight:bold;font-family:Arial, sans-serif")
			}
		} else {
			canvas.Text(int(x+node.Width/2), int(labelY+4), label, "text-anchor:middle;fill:#2c3e50;font-size:12;font-weight:bold;font-family:Arial, sans-serif")
		}

		if node.Version != "" {
			versionY := labelY + 24
			canvas.Text(int(x+node.Width/2), int(versionY), "v"+node.Version, "text-anchor:middle;fill:#6c757d;font-size:9;font-family:Arial, sans-serif")
		}

		cardX := int(x)
		cardWidth := node.Width
		cardSpacing := 8.0
		cardPadding := 8.0
		cardY := int(y + node.Height + cardSpacing)

		requestCount := len(node.RequestKeys)
		headerHeight := 18.0
		lineHeight := 16.0
		cardHeight := headerHeight + cardPadding + float64(requestCount)*lineHeight + cardPadding
		if requestCount == 0 {
			cardHeight = headerHeight + cardPadding*2 + 12
		}
		if requestCount > 5 {
			cardHeight = headerHeight + cardPadding + float64(5)*lineHeight + lineHeight + cardPadding
		}

		canvas.Roundrect(cardX, cardY, int(cardWidth), int(cardHeight), 4, 4, "fill:#e3f2fd;stroke:#90caf9;stroke-width:1")

		canvas.Text(cardX+int(cardWidth/2), cardY+int(headerHeight-2), "in", "text-anchor:middle;fill:#1976d2;font-size:9;font-weight:bold;font-family:Arial, sans-serif")

		textStartY := cardY + int(headerHeight+cardPadding)
		if requestCount == 0 {
			canvas.Text(cardX+int(cardPadding), textStartY+10, i18n.Msg("(no keys)"), "fill:#757575;font-size:8;font-style:italic;font-family:Arial, sans-serif")
		} else {
			maxKeys := 5
			for i, key := range node.RequestKeys {
				if i >= maxKeys {
					canvas.Text(cardX+int(cardPadding), textStartY+int(float64(i+1)*lineHeight), "...", "fill:#424242;font-size:8;font-family:Arial, sans-serif")
					break
				}
				displayKey := key
				if len(displayKey) > 15 {
					displayKey = displayKey[:12] + "..."
				}
				canvas.Text(cardX+int(cardPadding), textStartY+int(float64(i+1)*lineHeight), displayKey, "fill:#424242;font-size:8;font-family:Arial, sans-serif")
			}
		}

		responseCount := len(node.ResponseKeys)
		if responseCount > 0 {
			responseCardY := int(y + node.Height + cardHeight + cardSpacing*2)
			responseHeaderHeight := 18.0
			responseLineHeight := 16.0
			responseCardHeight := responseHeaderHeight + cardPadding + float64(responseCount)*responseLineHeight + cardPadding
			if responseCount > 5 {
				responseCardHeight = responseHeaderHeight + cardPadding + float64(5)*responseLineHeight + responseLineHeight + cardPadding
			}

			canvas.Roundrect(cardX, responseCardY, int(cardWidth), int(responseCardHeight), 4, 4, "fill:#e8f5e9;stroke:#81c784;stroke-width:1")

			canvas.Text(cardX+int(cardWidth/2), responseCardY+int(responseHeaderHeight-2), "out", "text-anchor:middle;fill:#2e7d32;font-size:9;font-weight:bold;font-family:Arial, sans-serif")

			responseTextStartY := responseCardY + int(responseHeaderHeight+cardPadding)
			maxKeys := 5
			for i, key := range node.ResponseKeys {
				if i >= maxKeys {
					canvas.Text(cardX+int(cardPadding), responseTextStartY+int(float64(i+1)*responseLineHeight), "...", "fill:#424242;font-size:8;font-family:Arial, sans-serif")
					break
				}
				displayKey := key
				if len(displayKey) > 15 {
					displayKey = displayKey[:12] + "..."
				}
				canvas.Text(cardX+int(cardPadding), responseTextStartY+int(float64(i+1)*responseLineHeight), displayKey, "fill:#424242;font-size:8;font-family:Arial, sans-serif")
			}
		}

		if node.IsCurrent {
			circleX := int(x + node.Width - 12)
			circleY := int(y + 12)
			canvas.Circle(circleX, circleY, 6, "fill:#27ae60")
		}

		if node.Duration != "" {
			durationX := int(x + node.Width - 6)
			durationY := int(y + node.Height - 6)
			canvas.Text(durationX, durationY, node.Duration, "text-anchor:end;fill:#95a5a6;font-size:8;font-family:Arial, sans-serif")
		}
	}

	for _, edge := range g.Edges {
		fromNode := findNodeByID(g.Nodes, edge.From)
		toNode := findNodeByID(g.Nodes, edge.To)
		if fromNode == nil || toNode == nil {
			continue
		}

		fromX := fromNode.X + fromNode.Width/2 + offsetX
		fromY := fromNode.Y + fromNode.Height/2 + offsetY
		toX := toNode.X + toNode.Width/2 + offsetX
		toY := toNode.Y + toNode.Height/2 + offsetY

		dx := toX - fromX
		dy := toY - fromY
		length := math.Sqrt(dx*dx + dy*dy)
		if length > 0 {
			unitX := dx / length
			unitY := dy / length
			fromX += unitX * (fromNode.Width / 2)
			fromY += unitY * (fromNode.Height / 2)
			toX -= unitX * (toNode.Width / 2)
			toY -= unitY * (toNode.Height / 2)
		}

		strokeColor := "#3498db"
		switch edge.Type {
		case "parallel":
			strokeColor = "#9b59b6"
		case "dependency":
			strokeColor = "#e74c3c"
		}

		canvas.Line(int(fromX), int(fromY), int(toX), int(toY), fmt.Sprintf("stroke:%s;stroke-width:1.5;marker-end:url(#arrowhead);opacity:0.7", strokeColor))
	}

	canvas.End()
	return buf.String()
}

// calculateBounds вычисляет границы графа.
func calculateBounds(g *graph) (minX, minY, maxX, maxY float64) {

	if len(g.Nodes) == 0 {
		return 0, 0, 400, 150
	}

	minX = g.Nodes[0].X
	minY = g.Nodes[0].Y
	maxX = g.Nodes[0].X + g.Nodes[0].Width
	maxY = g.Nodes[0].Y + g.Nodes[0].Height

	for _, node := range g.Nodes {
		if node.X < minX {
			minX = node.X
		}
		if node.Y < minY {
			minY = node.Y
		}

		nodeRight := node.X + node.Width
		if nodeRight > maxX {
			maxX = nodeRight
		}

		requestCount := len(node.RequestKeys)
		cardSpacing := 8.0
		cardPadding := 8.0
		headerHeight := 18.0
		lineHeight := 16.0

		requestCardHeight := headerHeight + cardPadding + float64(requestCount)*lineHeight + cardPadding
		if requestCount == 0 {
			requestCardHeight = headerHeight + cardPadding*2 + 12
		}
		if requestCount > 5 {
			requestCardHeight = headerHeight + cardPadding + float64(5)*lineHeight + lineHeight + cardPadding
		}

		totalCardHeight := requestCardHeight

		responseCount := len(node.ResponseKeys)
		if responseCount > 0 {
			responseLineHeight := 16.0
			responseCardHeight := headerHeight + cardPadding + float64(responseCount)*responseLineHeight + cardPadding
			if responseCount > 5 {
				responseCardHeight = headerHeight + cardPadding + float64(5)*responseLineHeight + responseLineHeight + cardPadding
			}
			totalCardHeight += responseCardHeight + cardSpacing
		}

		nodeBottom := node.Y + node.Height + totalCardHeight + cardSpacing
		if nodeBottom > maxY {
			maxY = nodeBottom
		}
	}

	return minX, minY, maxX, maxY
}

// getNodeStyle возвращает градиент и цвет для узла в зависимости от типа исполнителя.
func getNodeStyle(node *graphNode) (gradientID string, strokeColor string) {

	switch node.Type {
	case "wasm":
		return "wasmGrad", "#b3d9f2"
	case "native":
		return "nativeGrad", "#b8e6c1"
	case "command":
		return "commandGrad", "#ffd9b3"
	case "task":
		return "taskGrad", "#d4b3e3"
	case "transform":
		return "transformGrad", "#f5c2d6"
	case "plugin":
		return "wasmGrad", "#b3d9f2"
	default:
		return "defaultGrad", "#d0d0d0"
	}
}

// findNodeByID находит узел по ID.
func findNodeByID(nodes []*graphNode, id string) (node *graphNode) {

	for _, n := range nodes {
		if n.ID == id {
			return n
		}
	}
	return nil
}
