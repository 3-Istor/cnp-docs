#!/usr/bin/env bash
# Builds every view of the CNP system into docs/01-architecture/diagrams/ (drawio + png),
# dark and light. noodle refuses to build a view while boxes, labels or arrows collide.
#
# draw.io silently cuts off a PNG that is too big (platform-detailed at scale 2 is 203 megapixels
# and came out blank below half its height). Views whose scale-2 export exceeds MAX_PIXELS are
# exported at scale 1, and every PNG is then checked against the drawio: its last strip must
# hold content, or the script fails.
#
# VIEWS="a b" builds only those views. THEMES="light" builds only that theme (default: dark light).
#
# PNG_QUALITY is the pngquant min-max quality, 80-95 by default. draw.io exports PNGs
# several times larger than needed; "off" keeps them as exported.
set -euo pipefail

cd "$(dirname "$0")"
out_dir=../docs/01-architecture/diagrams
icons=../.noodle/icons
quality=${PNG_QUALITY:-80-95}
max_pixels=${MAX_PIXELS:-100000000}
border=20
if [[ $quality != off ]] && ! command -v pngquant >/dev/null; then
  echo "build-diagrams: pngquant not found; install it (nix shell nixpkgs#pngquant), or set PNG_QUALITY=off" >&2
  exit 1
fi
if ! command -v magick >/dev/null; then
  echo "build-diagrams: magick not found; install it (nix shell nixpkgs#imagemagick), it checks that the PNG is complete" >&2
  exit 1
fi
mkdir -p "$out_dir"

page_attr() { grep -o "$2=\"[0-9]*\"" "$1" | head -1 | tr -dc 0-9; }

# content_bottom <drawio>: lowest edge of the top-level cells, the background aside.
content_bottom() {
  awk '
    /<mxCell / { id = ""; top = / parent="1"/; if (match($0, / id="[^"]*"/)) id = substr($0, RSTART + 5, RLENGTH - 6) }
    /<mxGeometry / && top && id != "bg" && match($0, / y="[0-9.]+"/) {
      y = substr($0, RSTART + 4, RLENGTH - 5)
      if (match($0, / height="[0-9.]+"/)) { bottom = y + substr($0, RSTART + 9, RLENGTH - 10); if (bottom > max) max = bottom }
    }
    END { printf "%d", max }
  ' "$1"
}

# check_png_complete <drawio> <png> <scale>: the last 40 px of content must not be one flat
# colour in the PNG, since a card border or text ends there.
check_png_complete() {
  local drawio=$1 png=$2 scale=$3 bottom png_height strip_top deviation
  bottom=$(content_bottom "$drawio")
  png_height=$(magick identify -format %h "$png")
  strip_top=$(( (bottom - 40 + border) * scale ))
  deviation=0
  if (( strip_top + 40 * scale <= png_height )); then
    deviation=$(magick "$png" -crop "x$((40 * scale))+0+$strip_top" +repage -format '%[fx:standard_deviation]' info:)
  fi
  if [[ $deviation == 0 ]]; then
    echo "build-diagrams: $png is cut off: the drawio content ends at y=$bottom but the PNG, $png_height px high at scale $scale, is empty there. Lower MAX_PIXELS so that this view is exported at scale 1." >&2
    exit 1
  fi
}

for system in */; do
  system=${system%/}
  [[ -f $system/model.yaml ]] || continue
  for view in "$system"/views/*.yaml; do
    id=$(basename "$view" .yaml)
    [[ -z ${VIEWS:-} || " $VIEWS " == *" $id "* ]] || continue
    for theme in ${THEMES:-dark light}; do
      out=$out_dir/$id.$theme
      noodle render "$system" -view "$id" -icons "$icons" -theme "$theme" -o "$out.drawio"
      width=$(page_attr "$out.drawio" pageWidth)
      height=$(page_attr "$out.drawio" pageHeight)
      scale=2
      (( (width + 2 * border) * (height + 2 * border) * 4 > max_pixels )) && scale=1
      drawio -x -f png -s "$scale" --border "$border" -o "$out.png" "$out.drawio" >/dev/null 2>&1
      check_png_complete "$out.drawio" "$out.png" "$scale"
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
