package main

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

func (s *Spec) anchorRect(id string) (Rect, bool) {
	for _, n := range s.Nodes {
		if n.ID == id {
			return n.rect(), true
		}
	}
	for _, z := range s.Zones {
		if z.ID == id {
			return z.rect(), true
		}
	}
	return Rect{}, false
}

func borderSide(p Point, r Rect) string {
	const eps = 1.0
	inX := p.X() >= r.X-eps && p.X() <= r.X+r.W+eps
	inY := p.Y() >= r.Y-eps && p.Y() <= r.Y+r.H+eps
	switch {
	case math.Abs(p.X()-r.X) < eps && inY:
		return "left"
	case math.Abs(p.X()-r.X-r.W) < eps && inY:
		return "right"
	case math.Abs(p.Y()-r.Y) < eps && inX:
		return "top"
	case math.Abs(p.Y()-r.Y-r.H) < eps && inX:
		return "bottom"
	}
	return ""
}

type portBadge struct {
	ID, Node, Text, Side string
	Rect                 Rect
	Outer                Point
}

func portWidth(text string) float64 { return textWidth(text, portFontSize) + 2*portPadX }

// ports returns the badge each edge lands on. Edges reaching the same point of the same
// node with the same port share one badge.
func (s *Spec) ports() map[string]portBadge {
	byEdge := map[string]portBadge{}
	for _, e := range s.Edges {
		if e.Port == "" || len(e.Path) < 2 {
			continue
		}
		r, ok := s.anchorRect(e.To)
		if !ok {
			continue
		}
		p := e.Path[len(e.Path)-1]
		side := borderSide(p, r)
		w, h := portWidth(e.Port), portHeight
		b := portBadge{
			ID:   fmt.Sprintf("%s__port_%s_%.0f_%.0f", e.To, side, p.X(), p.Y()),
			Node: e.To, Text: e.Port, Side: side,
			Rect: Rect{p.X() - w/2, p.Y() - h/2, w, h},
		}
		switch side {
		case "left":
			b.Outer = Point{p.X() - w/2, p.Y()}
		case "right":
			b.Outer = Point{p.X() + w/2, p.Y()}
		case "top":
			b.Outer = Point{p.X(), p.Y() - h/2}
		default:
			b.Outer = Point{p.X(), p.Y() + h/2}
		}
		byEdge[e.ID] = b
	}
	return byEdge
}

// drawnPath is the path as rendered: it stops on the outer side of the port badge.
func drawnPath(e Edge, ports map[string]portBadge) []Point {
	p := append([]Point(nil), e.Path...)
	if b, ok := ports[e.ID]; ok {
		p[len(p)-1] = b.Outer
	}
	return p
}

func segLen(a, b Point) float64 { return math.Abs(b.X()-a.X()) + math.Abs(b.Y()-a.Y()) }

func onSegment(p, a, b Point) bool {
	const eps = 0.5
	return p.X() >= math.Min(a.X(), b.X())-eps && p.X() <= math.Max(a.X(), b.X())+eps &&
		p.Y() >= math.Min(a.Y(), b.Y())-eps && p.Y() <= math.Max(a.Y(), b.Y())+eps &&
		(math.Abs(a.X()-b.X()) < eps || math.Abs(a.Y()-b.Y()) < eps)
}

func labelAnchor(e Edge, path []Point) Point {
	if e.LabelAt != nil {
		return *e.LabelAt
	}
	best, bestLen := 0, -1.0
	for i := 0; i+1 < len(path); i++ {
		if l := segLen(path[i], path[i+1]); l > bestLen {
			best, bestLen = i, l
		}
	}
	a, b := path[best], path[best+1]
	return Point{(a.X() + b.X()) / 2, (a.Y() + b.Y()) / 2}
}

// labelFraction is where the anchor sits along the path, which is how draw.io positions edge labels.
func labelFraction(anchor Point, path []Point) float64 {
	total, before := 0.0, -1.0
	for i := 0; i+1 < len(path); i++ {
		a, b := path[i], path[i+1]
		if before < 0 && onSegment(anchor, a, b) {
			before = total + segLen(a, anchor)
		}
		total += segLen(a, b)
	}
	if total == 0 || before < 0 {
		return 0.5
	}
	return before / total
}

var (
	stepRe = regexp.MustCompile(`\[([0-9]{1,2}|[A-Z])\]`)
	boldRe = regexp.MustCompile(`\*\*(.+?)\*\*`)
	warnRe = regexp.MustCompile(`!!(.+?)!!`)
)

// plainText approximates the rendered width: a step badge is its text plus padding.
func plainText(s string) string {
	s = stepRe.ReplaceAllString(s, " $1 ")
	s = boldRe.ReplaceAllString(s, "$1")
	return warnRe.ReplaceAllString(s, "$1")
}

func textBox(cx, cy float64, lines []string, fontSize float64) Rect {
	w := 0.0
	for _, l := range lines {
		w = math.Max(w, textWidth(plainText(l), fontSize))
	}
	h := float64(len(lines))*fontSize*lineHeightEm + 2*labelPadY
	w += 2 * labelPadX
	return Rect{cx - w/2, cy - h/2, w, h}
}

func labelBox(e Edge, path []Point) (Rect, bool) {
	if e.Label == "" {
		return Rect{}, false
	}
	a := labelAnchor(e, path)
	return textBox(a.X()+e.LabelOffset.X(), a.Y()+e.LabelOffset.Y(), strings.Split(e.Label, "<br>"), edgeFontSize), true
}

func actorLabelBox(n Node) Rect {
	lines := append([]string{n.Title}, n.lines()...)
	b := textBox(n.X+n.W/2, 0, lines, titleFontSize)
	b.Y = n.Y + n.H + actorLabelGap - labelPadY
	return b
}

func nodeIconRect(n Node) Rect { return Rect{n.X + iconInset, n.Y + iconInset, iconSize, iconSize} }

func zoneTitleBox(z Zone) Rect {
	x := z.X + 12
	w := textWidth(z.Label, zoneFontSize) + 12
	if z.Icon != "" {
		w += zoneIconSize + 8
	}
	if z.Sub != "" {
		w += 2*zoneFontSize*charWidthEm + textWidth(z.Sub, subFontSize)
	}
	return Rect{x, z.Y + 4, w, math.Max(zoneFontSize*lineHeightEm, zoneIconSize) + 6}
}
