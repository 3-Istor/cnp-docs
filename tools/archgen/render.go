package main

import (
	"encoding/base64"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
)

type renderer struct {
	spec     *Spec
	th       *Theme
	iconsDir string
	icons    map[string]string
	ports    map[string]portBadge
	b        strings.Builder
}

// iconURI prefers <name>.<theme>.svg so a logo can swap to a variant that stays visible on the background.
func (r *renderer) iconURI(name string) (string, error) {
	if uri, ok := r.icons[name]; ok {
		return uri, nil
	}
	raw, err := os.ReadFile(filepath.Join(r.iconsDir, name+"."+r.th.Name+".svg"))
	if os.IsNotExist(err) {
		raw, err = os.ReadFile(filepath.Join(r.iconsDir, name+".svg"))
	}
	if err != nil {
		return "", fmt.Errorf("icon %q: %w", name, err)
	}
	// draw.io splits styles on ';', so its data URIs omit the ";base64" marker.
	uri := "data:image/svg+xml," + base64.StdEncoding.EncodeToString(raw)
	r.icons[name] = uri
	return uri, nil
}

func xmlAttr(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch c {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\n':
			b.WriteString("&#10;")
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

func style(parts ...string) string { return strings.Join(parts, ";") + ";" }

func font(size float64, color string) string {
	return fmt.Sprintf("fontFamily=%s;fontSize=%g;fontColor=%s", fontFamily, size, color)
}

func (r *renderer) stepBadge(step string) string {
	if step[0] >= 'A' && step[0] <= 'Z' {
		return fmt.Sprintf(`<span style="border:1px solid %s;color:%s;border-radius:3px;padding:0px 3px;font-weight:bold;">%s</span>`, r.th.StepFill, r.th.Title, step)
	}
	return fmt.Sprintf(`<span style="background-color:%s;color:%s;border-radius:3px;padding:0px 4px;font-weight:bold;">%s</span>`, r.th.StepFill, r.th.StepText, step)
}

// markup: [1] data-plane step (filled badge), [A] control-plane step (outlined badge), **bold**, !!warning!!.
func (r *renderer) markup(s string) string {
	parts := strings.Split(s, "<br>")
	for i, p := range parts {
		p = html.EscapeString(p)
		p = stepRe.ReplaceAllStringFunc(p, func(m string) string { return r.stepBadge(m[1 : len(m)-1]) })
		p = boldRe.ReplaceAllString(p, `<b><font color="`+r.th.Title+`">$1</font></b>`)
		p = warnRe.ReplaceAllString(p, `<b><font color="`+r.th.Warn+`">$1</font></b>`)
		parts[i] = p
	}
	return strings.Join(parts, "<br>")
}

func (r *renderer) vertex(id, parent, value, st string, x, y, w, h float64) {
	fmt.Fprintf(&r.b, `        <mxCell id="%s" value="%s" style="%s" vertex="1" parent="%s">
          <mxGeometry x="%g" y="%g" width="%g" height="%g" as="geometry" />
        </mxCell>
`, id, xmlAttr(value), xmlAttr(st), parent, x, y, w, h)
}

func (r *renderer) image(id, parent, icon string, x, y, size float64) error {
	uri, err := r.iconURI(icon)
	if err != nil {
		return err
	}
	r.vertex(id, parent, "", style("shape=image", "image="+uri, "imageAspect=1", "editable=0", "connectable=0", "movable=0", "resizable=0"), x, y, size, size)
	return nil
}

func (r *renderer) render() (string, error) {
	s := r.spec
	fmt.Fprintf(&r.b, `<mxfile host="archgen">
  <diagram id="%s-%s" name="%s (%s)">
    <mxGraphModel grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="0" pageScale="1" pageWidth="%g" pageHeight="%g" background="%s" math="0" shadow="0">
      <root>
        <mxCell id="0" />
        <mxCell id="1" parent="0" />
`, s.ID, r.th.Name, xmlAttr(s.Title), r.th.Name, s.Width, s.Height, r.th.Background)

	r.background()
	r.header()
	for _, z := range s.Zones {
		if err := r.zone(z); err != nil {
			return "", err
		}
	}
	// Edges before nodes so boxes and port badges sit on top of line ends.
	for _, e := range s.Edges {
		r.edge(e)
	}
	for _, n := range s.Nodes {
		if err := r.node(n); err != nil {
			return "", err
		}
	}
	r.portBadges()
	for _, n := range s.Notes {
		r.vertex(n.ID, "1", r.markup(n.Text), style("text", "html=1", "whiteSpace=wrap", "align=left", "verticalAlign=top", font(subFontSize, r.th.Muted)), n.X, n.Y, n.W, n.H)
	}
	for _, c := range s.Cards {
		r.card(c)
	}
	r.b.WriteString("      </root>\n    </mxGraphModel>\n  </diagram>\n</mxfile>\n")
	return r.b.String(), nil
}

func (r *renderer) background() {
	s := r.spec
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%g" height="%g"><defs><pattern id="g" width="40" height="40" patternUnits="userSpaceOnUse"><path d="M 40 0 L 0 0 0 40" fill="none" stroke="%s" stroke-width="0.5"/></pattern></defs><rect width="100%%" height="100%%" fill="%s"/><rect width="100%%" height="100%%" fill="url(#g)"/></svg>`,
		s.Width, s.Height, r.th.GridLine, r.th.Background)
	uri := "data:image/svg+xml," + base64.StdEncoding.EncodeToString([]byte(svg))
	r.vertex("bg", "1", "", style("shape=image", "image="+uri, "imageAspect=0", "locked=1", "editable=0", "movable=0", "resizable=0", "deletable=0"), 0, 0, s.Width, s.Height)
}

func (r *renderer) header() {
	s := r.spec
	r.vertex("header-dot", "1", "", style("ellipse", "fillColor="+r.th.Accent, "strokeColor=none"), 40, 44, 12, 12)
	title := fmt.Sprintf(`<b>%s</b>`, html.EscapeString(s.Title))
	r.vertex("header-title", "1", title, style("text", "html=1", "align=left", "verticalAlign=middle", font(24, r.th.Title)), 64, 30, s.Width-400, 40)
	var sub []string
	if s.Subtitle != "" {
		sub = append(sub, fmt.Sprintf(`<font color="%s" style="font-size:13px">%s</font>`, r.th.Text, r.markup(s.Subtitle)))
	}
	for _, m := range s.Meta {
		sub = append(sub, r.markup(m))
	}
	r.vertex("header-sub", "1", strings.Join(sub, "<br>"), style("text", "html=1", "align=left", "verticalAlign=top", font(11, r.th.Muted)), 64, 74, s.Width-400, 80)
}

func (r *renderer) zone(z Zone) error {
	color := r.th.Zones[z.Color]
	dash, width, opacity, arc := "4 4", 1.0, 5, 8
	if z.Kind == "region" {
		dash, width, opacity, arc = "8 4", 1.2, 3, 12
	}
	pad := 12.0
	if z.Icon != "" {
		pad += zoneIconSize + 8
	}
	v := fmt.Sprintf(`<b>%s</b>`, html.EscapeString(z.Label))
	if z.Sub != "" {
		v += fmt.Sprintf(`&nbsp;&nbsp;<font color="%s" style="font-size:%gpx">%s</font>`, r.th.Muted, subFontSize, html.EscapeString(z.Sub))
	}
	st := style("rounded=1", "absoluteArcSize=1", fmt.Sprintf("arcSize=%d", arc), "html=1",
		"fillColor="+color, fmt.Sprintf("fillOpacity=%d", opacity), "strokeColor="+color, fmt.Sprintf("strokeWidth=%g", width),
		"dashed=1", "dashPattern="+dash, "container=0", "collapsible=0")
	r.vertex(z.ID, "1", "", st, z.X, z.Y, z.W, z.H)
	// Title and icon share one row and are both centred on it; draw.io's own label
	// padding would otherwise leave the icon a few pixels above the text.
	title := zoneTitleBox(z)
	rowY := title.Y - z.Y
	r.vertex(z.ID+"__title", z.ID, v, style("text", "html=1", "align=left", "verticalAlign=middle", "spacing=0",
		fmt.Sprintf("spacingLeft=%g", pad-12), font(zoneFontSize, color), "movable=0", "resizable=0", "connectable=0"),
		12, rowY, title.W, title.H)
	if z.Icon != "" {
		return r.image(z.ID+"__icon", z.ID, z.Icon, 12, rowY+(title.H-zoneIconSize)/2, zoneIconSize)
	}
	return nil
}

func (r *renderer) nodeLabel(n Node) string {
	v := fmt.Sprintf(`<b><font color="%s" style="font-size:%gpx">%s</font></b>`, r.th.Title, titleFontSize, html.EscapeString(n.Title))
	if n.Badge != "" {
		v += fmt.Sprintf(` <b><font color="%s" style="font-size:%gpx">⚠ %s</font></b>`, r.th.Warn, subFontSize, html.EscapeString(n.Badge))
	}
	if n.Tech != "" {
		v += fmt.Sprintf(`<br><i><font color="%s" style="font-size:%gpx">%s</font></i>`, r.th.Text, subFontSize, html.EscapeString(n.Tech))
	}
	if n.Desc != "" {
		v += fmt.Sprintf(`<br><font color="%s" style="font-size:%gpx">%s</font>`, r.th.Muted, subFontSize, r.markup(n.Desc))
	}
	return v
}

func (r *renderer) node(n Node) error {
	k := r.th.Nodes[n.Kind]
	if n.Shape == "actor" {
		uri, err := r.iconURI(n.Icon)
		if err != nil {
			return err
		}
		v := fmt.Sprintf(`<b>%s</b>`, html.EscapeString(n.Title))
		for _, l := range n.lines() {
			v += fmt.Sprintf(`<br><font color="%s" style="font-size:%gpx">%s</font>`, r.th.Muted, subFontSize, html.EscapeString(l))
		}
		st := style("shape=image", "image="+uri, "imageAspect=1", "html=1", "verticalLabelPosition=bottom", "verticalAlign=top",
			"labelPosition=center", "align=center", fmt.Sprintf("spacingTop=%g", actorLabelGap), font(titleFontSize, r.th.Title))
		r.vertex(n.ID, "1", v, st, n.X, n.Y, n.W, n.H)
		return nil
	}
	shape := []string{"rounded=1", "absoluteArcSize=1", "arcSize=12"}
	top := 10.0
	if n.Shape == "cylinder" {
		shape = []string{"shape=cylinder3", "size=8", "boundedLbl=1", "backgroundOutline=1"}
		top = 18
	}
	pad := 10.0
	if n.Icon != "" {
		pad = textPadLeft
	}
	st := style(append(shape, "html=1", "whiteSpace=wrap", "fillColor="+k.Fill, "strokeColor="+k.Stroke, "strokeWidth=1.5",
		"align=left", "verticalAlign=top", fmt.Sprintf("spacingLeft=%g", pad), fmt.Sprintf("spacingTop=%g", top-4), "spacingRight=8",
		font(titleFontSize, r.th.Text))...)
	r.vertex(n.ID, "1", r.nodeLabel(n), st, n.X, n.Y, n.W, n.H)
	if n.Icon != "" {
		iconY := iconInset
		if n.Shape == "cylinder" {
			iconY += 8
		}
		// Child of the node so it moves with it when edited by hand in draw.io.
		return r.image(n.ID+"__icon", n.ID, n.Icon, iconInset, iconY, iconSize)
	}
	return nil
}

func (r *renderer) kindStrokeOf(id string) string {
	for _, n := range r.spec.Nodes {
		if n.ID == id {
			return r.th.Nodes[n.Kind].Stroke
		}
	}
	for _, z := range r.spec.Zones {
		if z.ID == id {
			return r.th.Zones[z.Color]
		}
	}
	return r.th.Muted
}

func (r *renderer) portBadges() {
	seen := map[string]bool{}
	for _, e := range r.spec.Edges {
		b, ok := r.ports[e.ID]
		if !ok || seen[b.ID] {
			continue
		}
		seen[b.ID] = true
		host, _ := r.spec.anchorRect(b.Node)
		st := style("rounded=1", "absoluteArcSize=1", "arcSize=6", "html=1", "fillColor="+r.kindStrokeOf(b.Node), "strokeColor="+r.th.Background,
			"strokeWidth=1", "align=center", "verticalAlign=middle", "fontStyle=1", font(portFontSize, r.th.PortText), "movable=0", "resizable=0")
		r.vertex(b.ID, b.Node, html.EscapeString(b.Text), st, b.Rect.X-host.X, b.Rect.Y-host.Y, b.Rect.W, b.Rect.H)
	}
}

var sideEntry = map[string][2]float64{"left": {0, 0.5}, "right": {1, 0.5}, "top": {0.5, 0}, "bottom": {0.5, 1}}

func (r *renderer) edge(e Edge) {
	k := r.th.Edges[e.Kind]
	src, _ := r.spec.anchorRect(e.From)
	path := drawnPath(e, r.ports)
	p0 := path[0]
	target := e.To
	var entry [2]float64
	if b, ok := r.ports[e.ID]; ok {
		target = b.ID
		entry = sideEntry[b.Side]
	} else {
		dst, _ := r.spec.anchorRect(e.To)
		pn := path[len(path)-1]
		entry = [2]float64{(pn.X() - dst.X) / dst.W, (pn.Y() - dst.Y) / dst.H}
	}
	endFill := "1"
	if k.EndArrow == "cross" || k.EndArrow == "open" {
		endFill = "0"
	}
	parts := []string{"edgeStyle=orthogonalEdgeStyle", "rounded=1", "orthogonalLoop=1", "html=1", "jumpStyle=arc", "jumpSize=10",
		"strokeColor=" + k.Stroke, fmt.Sprintf("strokeWidth=%g", k.Width),
		"endArrow=" + k.EndArrow, "endFill=" + endFill, "endSize=7",
		fmt.Sprintf("exitX=%.4f", (p0.X()-src.X)/src.W), fmt.Sprintf("exitY=%.4f", (p0.Y()-src.Y)/src.H), "exitPerimeter=0",
		fmt.Sprintf("entryX=%.4f", entry[0]), fmt.Sprintf("entryY=%.4f", entry[1]), "entryPerimeter=0",
		"labelBackgroundColor=" + r.th.Background, font(edgeFontSize, k.LabelColor)}
	if k.Dash != "" {
		parts = append(parts, "dashed=1", "dashPattern="+k.Dash)
	}
	if k.EndArrow == "cross" {
		parts = append(parts, "endSize=10")
	}
	frac := labelFraction(labelAnchor(e, path), path)
	fmt.Fprintf(&r.b, `        <mxCell id="%s" value="%s" style="%s" edge="1" parent="1" source="%s" target="%s">
          <mxGeometry x="%.4f" relative="1" as="geometry">
`, e.ID, xmlAttr(r.markup(e.Label)), xmlAttr(style(parts...)), e.From, target, frac*2-1)
	if e.LabelOffset != (Point{}) {
		fmt.Fprintf(&r.b, "            <mxPoint x=\"%g\" y=\"%g\" as=\"offset\" />\n", e.LabelOffset.X(), e.LabelOffset.Y())
	}
	if len(path) > 2 {
		r.b.WriteString("            <Array as=\"points\">\n")
		for _, p := range path[1 : len(path)-1] {
			fmt.Fprintf(&r.b, "              <mxPoint x=\"%g\" y=\"%g\" />\n", p.X(), p.Y())
		}
		r.b.WriteString("            </Array>\n")
	}
	r.b.WriteString("          </mxGeometry>\n        </mxCell>\n")
}

func (r *renderer) card(c Card) {
	color := r.th.Zones[c.Color]
	if color == "" {
		color = r.th.Muted
	}
	st := style("rounded=1", "absoluteArcSize=1", "arcSize=16", "html=1", "whiteSpace=wrap", "fillColor="+r.th.CardFill,
		"strokeColor="+r.th.CardStroke, "align=left", "verticalAlign=top", "spacingLeft=20", "spacingTop=14", "spacingRight=16",
		font(11, r.th.Text))
	head := fmt.Sprintf(`<font color="%s">●</font>&nbsp;<b><font color="%s" style="font-size:13px">%s</font></b><br><br>`, color, r.th.Title, html.EscapeString(c.Title))
	var lines []string
	for _, l := range c.Lines {
		if strings.HasPrefix(l, "[") || strings.HasPrefix(l, "!!") {
			lines = append(lines, r.markup(l))
		} else {
			lines = append(lines, "• "+r.markup(l))
		}
	}
	body := ""
	if !c.Legend {
		body = `<div style="line-height:1.6">` + strings.Join(lines, "<br>") + `</div>`
	}
	r.vertex(c.ID, "1", head+body, st, c.X, c.Y, c.W, c.H)
	if c.Legend {
		r.legend(c)
	}
}

func (r *renderer) legendText(id, text string, x, y, w float64) {
	r.vertex(id, "1", text, style("text", "html=1", "align=left", "verticalAlign=middle", font(10.5, r.th.Text)), x, y, w, 22)
}

func (r *renderer) legend(c Card) {
	usedNodes, usedEdges := map[string]bool{}, map[string]bool{}
	for _, n := range r.spec.Nodes {
		if n.Shape != "actor" {
			usedNodes[n.Kind] = true
		}
	}
	for _, e := range r.spec.Edges {
		usedEdges[e.Kind] = true
	}
	x, y := c.X+20, c.Y+52
	for _, kind := range nodeKindOrder {
		if !usedNodes[kind] {
			continue
		}
		k := r.th.Nodes[kind]
		r.vertex("legend-node-"+kind, "1", "", style("rounded=1", "absoluteArcSize=1", "arcSize=6", "fillColor="+k.Fill, "strokeColor="+k.Stroke, "strokeWidth=1.5"), x, y+4, 22, 14)
		r.legendText("legend-node-"+kind+"-text", k.Legend, x+32, y, c.W/2-60)
		y += 26
	}
	y += 8
	r.legendText("legend-steps", r.markup("[1] request step · [A] background step"), x, y, c.W/2-20)
	y += 26
	st := style("rounded=1", "absoluteArcSize=1", "arcSize=6", "html=1", "fillColor="+r.th.Nodes["backend"].Stroke, "strokeColor=none",
		"align=center", "verticalAlign=middle", "fontStyle=1", font(portFontSize, r.th.PortText))
	r.vertex("legend-port", "1", "TCP 80", st, x, y+3, portWidth("TCP 80"), portHeight)
	r.legendText("legend-port-text", "listening port on the server", x+52, y, c.W/2-70)

	x, y = c.X+c.W/2+10, c.Y+52
	for _, kind := range edgeKindOrder {
		if !usedEdges[kind] {
			continue
		}
		k := r.th.Edges[kind]
		endFill := "1"
		if k.EndArrow == "cross" || k.EndArrow == "open" {
			endFill = "0"
		}
		st := style("edgeStyle=none", "html=1", "strokeColor="+k.Stroke, fmt.Sprintf("strokeWidth=%g", k.Width), "endArrow="+k.EndArrow, "endFill="+endFill, "endSize=6")
		if k.Dash != "" {
			st += "dashed=1;dashPattern=" + k.Dash + ";"
		}
		fmt.Fprintf(&r.b, `        <mxCell id="legend-edge-%s" value="" style="%s" edge="1" parent="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="%g" y="%g" as="sourcePoint" />
            <mxPoint x="%g" y="%g" as="targetPoint" />
          </mxGeometry>
        </mxCell>
`, kind, xmlAttr(st), x, y+11, x+40, y+11)
		r.legendText("legend-edge-"+kind+"-text", k.Legend, x+52, y, c.W/2-72)
		y += 26
	}
	notes := fmt.Sprintf(`Large dashed frame: infra perimeter · small: functional group<br><font color="%s"><b>⚠ Gx</b></font>: docs and code disagree, see "Gaps"`, r.th.Warn)
	r.vertex(c.ID+"-notes", "1", notes, style("text", "html=1", "align=left", "verticalAlign=top", "whiteSpace=wrap", font(10.5, r.th.Muted)), c.X+20, c.Y+c.H-50, c.W-40, 40)
}
