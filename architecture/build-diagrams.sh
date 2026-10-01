#!/usr/bin/env bash
# Builds every view of the CNP system into docs/01-architecture/diagrams/ (drawio + png),
# dark and light. noodle refuses to build a view while boxes, labels or arrows collide.
#
# PNG_QUALITY is the pngquant min-max quality, 80-95 by default. draw.io exports PNGs
# several times larger than needed; "off" keeps them as exported.
set -euo pipefail

cd "$(dirname "$0")"
out_dir=../docs/01-architecture/diagrams
icons=../.noodle/icons
quality=${PNG_QUALITY:-80-95}
if [[ $quality != off ]] && ! command -v pngquant >/dev/null; then
  echo "build-diagrams: pngquant not found; install it (nix shell nixpkgs#pngquant), or set PNG_QUALITY=off" >&2
  exit 1
fi
mkdir -p "$out_dir"

for system in */; do
  system=${system%/}
  [[ -f $system/model.yaml ]] || continue
  for view in "$system"/views/*.yaml; do
    id=$(basename "$view" .yaml)
    # The view itself, then one output per lens it declares (noodle -lens, ADR-0020).
    lenses=$(sed -n 's/^  - {id: \([a-z0-9-]*\), subject:.*/\1/p' "$view")
    for lens in "" $lenses; do
      name=$id
      args=(-view "$id")
      if [[ -n $lens ]]; then
        name=$id.$lens
        args+=(-lens "$lens")
      fi
      for theme in dark light; do
        out=$out_dir/$name.$theme
        noodle render "$system" "${args[@]}" -icons "$icons" -theme "$theme" -o "$out.drawio"
        drawio -x -f png -s 2 --border 20 -o "$out.png" "$out.drawio" >/dev/null 2>&1
        if [[ $quality != off ]]; then
          # 98 and 99: the result would be larger or below the minimum quality; keep the export.
          status=0
          pngquant --quality "$quality" --speed 1 --skip-if-larger --force --output "$out.png" "$out.png" || status=$?
          [[ $status == 0 || $status == 98 || $status == 99 ]] || exit "$status"
        fi
        echo "built $out"
      done
    done
  done
done
