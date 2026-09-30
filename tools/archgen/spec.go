package main

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Point [2]float64

func (p Point) X() float64 { return p[0] }
func (p Point) Y() float64 { return p[1] }

type Rect struct{ X, Y, W, H float64 }

func (r Rect) inflate(d float64) Rect { return Rect{r.X - d, r.Y - d, r.W + 2*d, r.H + 2*d} }

func (r Rect) intersects(o Rect) bool {
	return r.X < o.X+o.W && o.X < r.X+r.W && r.Y < o.Y+o.H && o.Y < r.Y+r.H
}

func (r Rect) contains(o Rect) bool {
	return o.X >= r.X && o.Y >= r.Y && o.X+o.W <= r.X+r.W && o.Y+o.H <= r.Y+r.H
}

type Spec struct {
	ID       string   `yaml:"id"`
	Title    string   `yaml:"title"`
	Subtitle string   `yaml:"subtitle"`
	Meta     []string `yaml:"meta"`
	Width    float64  `yaml:"width"`
	Height   float64  `yaml:"height"`
	Zones    []Zone   `yaml:"zones"`
	Nodes    []Node   `yaml:"nodes"`
	Edges    []Edge   `yaml:"edges"`
	Notes    []Note   `yaml:"notes"`
	Cards    []Card   `yaml:"cards"`
}

// Zone is a dashed boundary: "region" for an infra or trust perimeter, "group" for a functional category inside one.
type Zone struct {
	ID    string  `yaml:"id"`
	Kind  string  `yaml:"kind"`
	Label string  `yaml:"label"`
	Sub   string  `yaml:"sub"`
	Color string  `yaml:"color"`
	Icon  string  `yaml:"icon"`
	X     float64 `yaml:"x"`
	Y     float64 `yaml:"y"`
	W     float64 `yaml:"w"`
	H     float64 `yaml:"h"`
}

func (z Zone) rect() Rect { return Rect{z.X, z.Y, z.W, z.H} }

// Node text follows the C4 triptych: name, [technology], one-line responsibility.
type Node struct {
	ID    string  `yaml:"id"`
	Kind  string  `yaml:"kind"`
	Shape string  `yaml:"shape"` // box (default), cylinder, actor
	Icon  string  `yaml:"icon"`
	Title string  `yaml:"title"`
	Tech  string  `yaml:"tech"`
	Desc  string  `yaml:"desc"`
	Badge string  `yaml:"badge"`
	X     float64 `yaml:"x"`
	Y     float64 `yaml:"y"`
	W     float64 `yaml:"w"`
	H     float64 `yaml:"h"`
}

func (n Node) rect() Rect { return Rect{n.X, n.Y, n.W, n.H} }

func (n Node) lines() []string {
	var l []string
	if n.Tech != "" {
		l = append(l, n.Tech)
	}
	if n.Desc != "" {
		l = append(l, n.Desc)
	}
	return l
}

// Edge is drawn from the side that opens the connection to the side that listens.
// Port names the listening port on the target and becomes a badge on its border.
type Edge struct {
	ID          string  `yaml:"id"`
	From        string  `yaml:"from"`
	To          string  `yaml:"to"`
	Kind        string  `yaml:"kind"`
	Label       string  `yaml:"label"`
	Port        string  `yaml:"port"`
	AgainstFlow bool    `yaml:"against_flow"` // outbound connection drawn against the reading direction, e.g. a tunnel
	Path        []Point `yaml:"path"`
	LabelAt     *Point  `yaml:"label_at"`
	LabelOffset Point   `yaml:"label_offset"`
}

type Note struct {
	ID   string  `yaml:"id"`
	Text string  `yaml:"text"`
	X    float64 `yaml:"x"`
	Y    float64 `yaml:"y"`
	W    float64 `yaml:"w"`
	H    float64 `yaml:"h"`
}

func (n Note) rect() Rect { return Rect{n.X, n.Y, n.W, n.H} }

type Card struct {
	ID     string   `yaml:"id"`
	Title  string   `yaml:"title"`
	Color  string   `yaml:"color"`
	Legend bool     `yaml:"legend"`
	Lines  []string `yaml:"lines"`
	X      float64  `yaml:"x"`
	Y      float64  `yaml:"y"`
	W      float64  `yaml:"w"`
	H      float64  `yaml:"h"`
}

func (c Card) rect() Rect { return Rect{c.X, c.Y, c.W, c.H} }

func loadSpec(path string) (*Spec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loadSpec: %w", err)
	}
	var s Spec
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("loadSpec %s: %w", path, err)
	}
	return &s, nil
}
