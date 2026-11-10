#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-only
set -euo pipefail
cd -- "$(dirname -- "$0")/.."
: "${QMLSCENE:=qmlscene}"
command -v "$QMLSCENE" >/dev/null || { echo 'Set QMLSCENE to a Qt 6 qmlscene executable.' >&2; exit 1; }
mkdir -p docs/images
stage=$(mktemp -d examples/.screenshots.XXXXXX)
trap 'rm -rf -- "$stage"' EXIT
export QT_QPA_PLATFORM=offscreen QT_QPA_PLATFORMTHEME=generic QT_QUICK_BACKEND=software
for screen in week day tasks keyboard; do
  view=$screen
  keyboard=false
  if [[ "$screen" == keyboard ]]; then view=day; keyboard=true; fi
  cat > "$stage/capture.qml" <<QML
import QtQuick
import ".."
Demo { demoView: "$view"; capturePath: "docs/images/$screen.png"; keyboardPreview: $keyboard }
QML
  "$QMLSCENE" -I tests/qml/imports "$stage/capture.qml"
done
printf 'Rendered four screenshots from fictional fixtures.\n'
