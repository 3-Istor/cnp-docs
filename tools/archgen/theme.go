package main

import "fmt"

// Palettes are Tailwind CSS v3 (the one Cocoon-AI uses). Dark: 400 strokes on slate-950,
// fills pre-composited so arrows never show through a box. Light: 600 strokes, 50 fills,
// 700 for coloured text so every text pair clears WCAG AA (4.5:1).
type Theme struct {
	Name                             string
	Background, GridLine             string
	CardFill, CardStroke             string
	Title, Text, Muted, Warn, Accent string
	StepFill, StepText               string
	PortText                         string
	Nodes                            map[string]nodeKind
	Zones                            map[string]string
	Edges                            map[string]edgeKind
}

type nodeKind struct{ Stroke, Fill, Legend string }

type edgeKind struct {
	Stroke, LabelColor, Dash, EndArrow, Legend string
	Width                                      float64
}

const (
	fontFamily    = "JetBrains Mono,Noto Sans Mono,DejaVu Sans Mono,monospace"
	titleFontSize = 12.0
	subFontSize   = 10.0
	edgeFontSize  = 10.0
	zoneFontSize  = 11.0
	portFontSize  = 9.0
	charWidthEm   = 0.6
	lineHeightEm  = 1.25

	iconSize      = 28.0
	iconInset     = 12.0
	textPadLeft   = iconInset + iconSize + 10
	zoneIconSize  = 18.0
	actorLabelGap = 6.0
	labelPadX     = 4.0
	labelPadY     = 2.0
	portPadX      = 4.0
	portHeight    = 16.0

	minNodeGap     = 24.0
	minZoneGap     = 60.0
	obstacleMargin = 4.0
)

var nodeKindOrder = []string{"frontend", "backend", "database", "cloud", "security", "bus", "external"}
var edgeKindOrder = []string{"flow", "auth", "tunnel", "async", "blocked", "link"}

var legendNodes = map[string]string{
	"frontend": "client / frontend",
	"backend":  "service / routing",
	"database": "data / secrets",
	"cloud":    "edge / tunnel / cloud",
	"security": "security / identity",
	"bus":      "bus / messaging",
	"external": "external / out of scope",
}

var legendEdges = map[string]string{
	"flow":    "connection, client → server",
	"auth":    "authentication connection",
	"tunnel":  "outbound tunnel, opened in advance",
	"async":   "background sync",
	"blocked": "must not happen",
	"link":    "object reference, not traffic",
}

func nodes(pairs ...string) map[string]nodeKind {
	m := map[string]nodeKind{}
	for i := 0; i < len(pairs); i += 3 {
		m[pairs[i]] = nodeKind{Stroke: pairs[i+1], Fill: pairs[i+2], Legend: legendNodes[pairs[i]]}
	}
	return m
}

func edges(pairs map[string][2]string, blockedEnd string) map[string]edgeKind {
	dash := map[string]string{"auth": "6 4", "tunnel": "3 3", "async": "6 4", "blocked": "6 4", "link": "2 3"}
	width := map[string]float64{"blocked": 1.8, "link": 1.2}
	m := map[string]edgeKind{}
	for k, c := range pairs {
		w := width[k]
		if w == 0 {
			w = 1.5
		}
		end := "blockThin"
		if k == "blocked" {
			end = blockedEnd
		}
		if k == "link" {
			end = "open"
		}
		m[k] = edgeKind{Stroke: c[0], LabelColor: c[1], Dash: dash[k], EndArrow: end, Legend: legendEdges[k], Width: w}
	}
	return m
}

var themes = map[string]*Theme{
	"dark": {
		Name: "dark", Background: "#020617", GridLine: "#1e293b",
		CardFill: "#0b1222", CardStroke: "#1e293b",
		Title: "#f8fafc", Text: "#e2e8f0", Muted: "#94a3b8", Warn: "#fb923c", Accent: "#34d399",
		StepFill: "#f8fafc", StepText: "#020617", PortText: "#020617",
		Nodes: nodes(
			"frontend", "#22d3ee", "#0c2234",
			"backend", "#34d399", "#0b2d31",
			"database", "#a78bfa", "#271955",
			"cloud", "#fbbf24", "#2f2022",
			"security", "#fb7185", "#3f152f",
			"bus", "#fb923c", "#563c2f",
			"external", "#94a3b8", "#172033",
		),
		Zones: map[string]string{"cyan": "#22d3ee", "emerald": "#34d399", "violet": "#a78bfa", "amber": "#fbbf24",
			"rose": "#fb7185", "orange": "#fb923c", "slate": "#94a3b8", "indigo": "#818cf8", "sky": "#38bdf8"},
		Edges: edges(map[string][2]string{
			"flow": {"#94a3b8", "#e2e8f0"}, "auth": {"#fb7185", "#fda4af"}, "tunnel": {"#fbbf24", "#fcd34d"},
			"async": {"#a78bfa", "#c4b5fd"}, "blocked": {"#f87171", "#fca5a5"}, "link": {"#64748b", "#94a3b8"},
		}, "cross"),
	},
	"light": {
		Name: "light", Background: "#f8fafc", GridLine: "#e2e8f0",
		CardFill: "#ffffff", CardStroke: "#e2e8f0",
		Title: "#0f172a", Text: "#1e293b", Muted: "#475569", Warn: "#c2410c", Accent: "#059669",
		StepFill: "#0f172a", StepText: "#ffffff", PortText: "#ffffff",
		Nodes: nodes(
			"frontend", "#0891b2", "#ecfeff",
			"backend", "#059669", "#ecfdf5",
			"database", "#7c3aed", "#f5f3ff",
			"cloud", "#d97706", "#fffbeb",
			"security", "#e11d48", "#fff1f2",
			"bus", "#ea580c", "#fff7ed",
			"external", "#64748b", "#f1f5f9",
		),
		Zones: map[string]string{"cyan": "#0891b2", "emerald": "#059669", "violet": "#7c3aed", "amber": "#d97706",
			"rose": "#e11d48", "orange": "#ea580c", "slate": "#64748b", "indigo": "#4f46e5", "sky": "#0284c7"},
		Edges: edges(map[string][2]string{
			"flow": {"#64748b", "#1e293b"}, "auth": {"#e11d48", "#be123c"}, "tunnel": {"#d97706", "#b45309"},
			"async": {"#7c3aed", "#6d28d9"}, "blocked": {"#dc2626", "#b91c1c"}, "link": {"#94a3b8", "#475569"},
		}, "cross"),
	},
}

func themeByName(name string) (*Theme, error) {
	t, ok := themes[name]
	if !ok {
		return nil, fmt.Errorf("unknown theme %q", name)
	}
	return t, nil
}

func textWidth(s string, fontSize float64) float64 {
	return float64(len([]rune(s))) * fontSize * charWidthEm
}
