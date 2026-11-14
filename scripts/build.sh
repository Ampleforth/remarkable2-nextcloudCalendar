#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-only
set -euo pipefail
cd -- "$(dirname -- "$0")/.."
: "${GO:=go}"
: "${RCC:=/usr/lib/qt6/rcc}"
"$GO" test ./...
target_arch=${CALENDAR_GOARCH:-arm}
case "$target_arch" in
  arm) app=dist/nextcloud-calendar; machine=armv7l; logo=qml/calendar.svg ;;
  arm64) app=dist/nextcloud-calendar-arm64; machine=aarch64; logo=qml/calendar-color.svg ;;
  *) echo 'CALENDAR_GOARCH must be arm or arm64.' >&2; exit 2 ;;
esac
mkdir -p "$app/backend"
if [[ "$target_arch" == arm ]]; then
  CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 "$GO" build -buildvcs=false -trimpath -ldflags='-s -w' -o "$app/backend/entry" .
else
  CGO_ENABLED=0 GOOS=linux GOARCH=arm64 "$GO" build -buildvcs=false -trimpath -ldflags='-s -w' -o "$app/backend/entry" .
fi
"$RCC" --binary --format-version 3 --no-compress packaging/application.qrc -o "$app"/resources.rcc
mkdir -p "$app"/xovi
cp packaging/xovi/*.qmd "$app"/xovi/
cp packaging/manifest.json "$app"/
cp LICENSE LICENSE.upstream.md THIRD_PARTY_NOTICES.md "$app"/
mkdir -p "$app"/licenses
for notice in licenses/*; do install -m 644 "$notice" "$app"/licenses/; done
printf 'Corresponding source is included in the release archive under source/.\nProject: https://github.com/Ampleforth/remarkable2-nextcloudCalendar\n' > "$app"/SOURCE.txt
printf '%s\n' "$machine" > "$app/architecture"
cp VERSION "$app/VERSION"
rsvg-convert -w 192 -h 192 "$logo" -o "$app"/icon.png
printf 'Built %s/ (%s).\n' "$app" "$machine"
