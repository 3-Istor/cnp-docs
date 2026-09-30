package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type obstacle struct {
	name  string
	rect  Rect
	owner string // edge id owning this label or port, node id for a node
	node  bool
}

type finding struct{ where, msg string }

type linter struct {
	s        *Spec
	ports    map[string]portBadge
	findings []finding
}

func (l *linter) errf(where, format string, args ...any) {
	l.findings = append(l.findings, finding{where, fmt.Sprintf(format, args...)})
}

func segRect(a, b Point) Rect {
	return Rect{math.Min(a.X(), b.X()), math.Min(a.Y(), b.Y()), math.Abs(b.X() - a.X()), math.Abs(b.Y() - a.Y())}
}

// segHits treats the segment as a zero-width rectangle so a line grazing a box counts as a hit.
func segHits(a, b Point, r Rect) bool {
	s := segRect(a, b)
	return s.X <= r.X+r.W && r.X <= s.X+s.W && s.Y <= r.Y+r.H && r.Y <= s.Y+s.H
}

func lint(s *Spec) []finding {
	l := &linter{s: s, ports: s.ports()}
	l.checkReferences()
	l.checkNodes()
	l.checkZones()
	l.checkPorts()
	l.checkEdges()
	sort.SliceStable(l.findings, func(i, j int) bool { return l.findings[i].where < l.findings[j].where })
	return l.findings
}

func (l *linter) checkReferences() {
	seen := map[string]bool{}
	check := func(id string) {
		if seen[id] {
			l.errf(id, "duplicate id")
		}
		seen[id] = true
	}
	for _, z := range l.s.Zones {
		check(z.ID)
		if _, ok := themes["dark"].Zones[z.Color]; !ok {
			l.errf(z.ID, "unknown zone color %q", z.Color)
		}
	}
	for _, n := range l.s.Nodes {
		check(n.ID)
		if _, ok := themes["dark"].Nodes[n.Kind]; !ok {
			l.errf(n.ID, "unknown node kind %q", n.Kind)
		}
	}
	for _, e := range l.s.Edges {
		check(e.ID)
		if _, ok := themes["dark"].Edges[e.Kind]; !ok {
			l.errf(e.ID, "unknown edge kind %q", e.Kind)
		}
		for _, end := range []string{e.From, e.To} {
			if _, ok := l.s.anchorRect(end); !ok {
				l.errf(e.ID, "unknown endpoint %q", end)
			}
		}
		if len(e.Path) < 2 {
			l.errf(e.ID, "path needs at least 2 points")
		}
	}
}

func (l *linter) checkNodes() {
	for i, n := range l.s.Nodes {
		if n.Shape != "actor" {
			avail := n.W - 10 - 8
			if n.Icon != "" {
				avail = n.W - textPadLeft - 8
			}
			title := n.Title
			if n.Badge != "" {
				title += " ⚠ " + n.Badge
			}
			if w := textWidth(title, titleFontSize); w > avail {
				l.errf(n.ID, "title overflows: %.0fpx text in %.0fpx", w, avail)
			}
			for _, line := range n.lines() {
				if w := textWidth(plainText(line), subFontSize); w > avail {
					l.errf(n.ID, "%q overflows: %.0fpx text in %.0fpx", line, w, avail)
				}
			}
			need := 10 + titleFontSize*lineHeightEm + float64(len(n.lines()))*subFontSize*lineHeightEm + 10
			if n.Shape == "cylinder" {
				need += 16
			}
			if need > n.H {
				l.errf(n.ID, "text needs %.0fpx height, box is %.0fpx", need, n.H)
			}
		}
		for _, o := range l.s.Nodes[i+1:] {
			if n.rect().inflate(minNodeGap / 2).intersects(o.rect().inflate(minNodeGap / 2)) {
				l.errf(n.ID, "closer than %.0fpx to %s", minNodeGap, o.ID)
			}
		}
		for _, z := range l.s.Zones {
			if n.rect().intersects(zoneTitleBox(z)) {
				l.errf(n.ID, "covers title of zone %s", z.ID)
			}
			if n.rect().intersects(z.rect()) && !z.rect().contains(n.rect()) {
				l.errf(n.ID, "straddles border of zone %s", z.ID)
			}
		}
	}
}

func (l *linter) checkZones() {
	for i, a := range l.s.Zones {
		for _, b := range l.s.Zones[i+1:] {
			ra, rb := a.rect(), b.rect()
			if ra.contains(rb) || rb.contains(ra) {
				continue
			}
			if ra.intersects(rb) {
				l.errf(a.ID, "overlaps zone %s", b.ID)
				continue
			}
			if ra.inflate(minZoneGap / 2).intersects(rb.inflate(minZoneGap / 2)) {
				l.errf(a.ID, "closer than %.0fpx to zone %s", minZoneGap, b.ID)
			}
		}
	}
}

func (l *linter) checkPorts() {
	seen := map[string]bool{}
	for _, e := range l.s.Edges {
		b, ok := l.ports[e.ID]
		if !ok || seen[b.ID] {
			continue
		}
		seen[b.ID] = true
		if b.Side == "" {
			l.errf(e.ID, "port %q: last point is not on a border of %s", e.Port, e.To)
			continue
		}
		for _, n := range l.s.Nodes {
			if n.ID == b.Node && n.Icon != "" && b.Rect.intersects(nodeIconRect(n).inflate(2)) {
				l.errf(e.ID, "port badge %q covers the icon of %s", b.Text, n.ID)
			}
		}
	}
}

func (l *linter) obstacles() []obstacle {
	var obs []obstacle
	for _, n := range l.s.Nodes {
		obs = append(obs, obstacle{name: "node " + n.ID, rect: n.rect(), owner: n.ID, node: true})
		if n.Shape == "actor" {
			obs = append(obs, obstacle{name: "label of " + n.ID, rect: actorLabelBox(n), owner: n.ID})
		}
	}
	for _, z := range l.s.Zones {
		obs = append(obs, obstacle{name: "title of zone " + z.ID, rect: zoneTitleBox(z)})
	}
	for _, nt := range l.s.Notes {
		obs = append(obs, obstacle{name: "note " + nt.ID, rect: nt.rect()})
	}
	for _, c := range l.s.Cards {
		obs = append(obs, obstacle{name: "card " + c.ID, rect: c.rect()})
	}
	seen := map[string]bool{}
	for _, e := range l.s.Edges {
		if b, ok := labelBox(e, drawnPath(e, l.ports)); ok {
			obs = append(obs, obstacle{name: "label of " + e.ID, rect: b, owner: e.ID})
		}
		if b, ok := l.ports[e.ID]; ok && !seen[b.ID] {
			seen[b.ID] = true
			obs = append(obs, obstacle{name: "port " + b.ID, rect: b.Rect, owner: b.ID})
		}
	}
	return obs
}

var connectionKinds = map[string]bool{"flow": true, "auth": true, "tunnel": true, "async": true, "blocked": true}

func (l *linter) checkEdges() {
	obs := l.obstacles()
	for _, e := range l.s.Edges {
		if len(e.Path) < 2 {
			continue
		}
		src, _ := l.s.anchorRect(e.From)
		dst, _ := l.s.anchorRect(e.To)
		out, in := borderSide(e.Path[0], src), borderSide(e.Path[len(e.Path)-1], dst)
		if out == "" {
			l.errf(e.ID, "first point %v is not on the border of %s", e.Path[0], e.From)
		}
		if in == "" {
			l.errf(e.ID, "last point %v is not on the border of %s", e.Path[len(e.Path)-1], e.To)
		}
		if connectionKinds[e.Kind] && !e.AgainstFlow {
			if out != "" && out != "right" && out != "bottom" {
				l.errf(e.ID, "egress leaves %s from the %s side, want right or bottom", e.From, out)
			}
			if in != "" && in != "left" && in != "top" {
				l.errf(e.ID, "ingress enters %s from the %s side, want left or top", e.To, in)
			}
		}
		path := drawnPath(e, l.ports)
		if e.LabelAt != nil {
			on := false
			for i := 0; i+1 < len(path); i++ {
				on = on || onSegment(*e.LabelAt, path[i], path[i+1])
			}
			if !on {
				l.errf(e.ID, "label_at %v is not on the drawn path", *e.LabelAt)
			}
		}
		own := map[string]bool{e.ID: true}
		if b, ok := l.ports[e.ID]; ok {
			own[b.ID] = true
		}
		last := len(path) - 2
		for i := 0; i <= last; i++ {
			a, b := path[i], path[i+1]
			if a.X() != b.X() && a.Y() != b.Y() {
				l.errf(e.ID, "segment %d is diagonal", i)
			}
			for _, o := range obs {
				if own[o.owner] {
					continue
				}
				r := o.rect
				if o.node {
					if (i == 0 && o.owner == e.From) || (i == last && o.owner == e.To) {
						continue
					}
					r = r.inflate(obstacleMargin)
				}
				if segHits(a, b, r) {
					l.errf(e.ID, "segment %d crosses %s", i, o.name)
				}
			}
		}
		if lb, ok := labelBox(e, path); ok {
			for _, o := range obs {
				if o.owner == e.ID || strings.HasPrefix(o.name, "card ") {
					continue
				}
				if lb.intersects(o.rect) {
					l.errf(e.ID, "label overlaps %s", o.name)
				}
			}
		}
	}
	l.checkCollinear()
	l.checkSteps()
}

// checkSteps keeps numbering checkable: one connection carries at most one step, and a
// step appears on one connection only. A flow that reuses connections belongs in a
// sequence diagram.
func (l *linter) checkSteps() {
	owner := map[string]string{}
	for _, e := range l.s.Edges {
		steps := stepRe.FindAllStringSubmatch(e.Label, -1)
		if len(steps) > 1 {
			l.errf(e.ID, "carries %d steps, want at most one", len(steps))
		}
		for _, m := range steps {
			if prev, ok := owner[m[1]]; ok {
				l.errf(e.ID, "step [%s] already used on %s", m[1], prev)
				continue
			}
			owner[m[1]] = e.ID
		}
	}
}

func (l *linter) checkCollinear() {
	type seg struct {
		edge string
		a, b Point
	}
	var segs []seg
	for _, e := range l.s.Edges {
		p := drawnPath(e, l.ports)
		for i := 0; i+1 < len(p); i++ {
			segs = append(segs, seg{e.ID, p[i], p[i+1]})
		}
	}
	for i, s1 := range segs {
		for _, s2 := range segs[i+1:] {
			if s1.edge == s2.edge {
				continue
			}
			r1, r2 := segRect(s1.a, s1.b), segRect(s2.a, s2.b)
			sameH := r1.H == 0 && r2.H == 0 && math.Abs(r1.Y-r2.Y) < 6
			sameV := r1.W == 0 && r2.W == 0 && math.Abs(r1.X-r2.X) < 6
			if sameH && r1.X < r2.X+r2.W && r2.X < r1.X+r1.W {
				l.errf(s1.edge, "runs on top of %s (horizontal, y≈%.0f)", s2.edge, r1.Y)
			}
			if sameV && r1.Y < r2.Y+r2.H && r2.Y < r1.Y+r1.H {
				l.errf(s1.edge, "runs on top of %s (vertical, x≈%.0f)", s2.edge, r1.X)
			}
		}
	}
}
