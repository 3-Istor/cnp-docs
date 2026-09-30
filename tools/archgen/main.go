// archgen turns an architecture spec (YAML) into a draw.io file in the CNP house style,
// refusing to emit a diagram whose lines cross boxes or whose labels collide.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	out := flag.String("o", "", "output .drawio path (lint only when empty)")
	icons := flag.String("icons", "", "directory holding <name>.svg icons, optionally <name>.<theme>.svg")
	themeName := flag.String("theme", "dark", "dark or light")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: archgen [-o out.drawio] [-theme dark|light] -icons DIR spec.yaml")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *out, *icons, *themeName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(specPath, out, iconsDir, themeName string) error {
	th, err := themeByName(themeName)
	if err != nil {
		return err
	}
	spec, err := loadSpec(specPath)
	if err != nil {
		return err
	}
	findings := lint(spec)
	for _, f := range findings {
		fmt.Fprintf(os.Stderr, "lint: %-28s %s\n", f.where, f.msg)
	}
	if len(findings) > 0 {
		return fmt.Errorf("%d lint finding(s)", len(findings))
	}
	if out == "" {
		return nil
	}
	r := &renderer{spec: spec, th: th, iconsDir: iconsDir, icons: map[string]string{}, ports: spec.ports()}
	xml, err := r.render()
	if err != nil {
		return err
	}
	return os.WriteFile(out, []byte(xml), 0o644)
}
