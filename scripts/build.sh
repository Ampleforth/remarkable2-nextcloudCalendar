#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-only
set -euo pipefail
cd -- "$(dirname -- "$0")/.."
: "${GO:=go}"
: "${RCC:=/usr/lib/qt6/rcc}"
"$GO" test ./...
mkdir -p dist/nextcloud-calendar/backend
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 "$GO" build -buildvcs=false -trimpath -ldflags='-s -w' -o dist/nextcloud-calendar/backend/entry .
"$RCC" --binary --format-version 3 --no-compress packaging/application.qrc -o dist/nextcloud-calendar/resources.rcc
cp packaging/manifest.json dist/nextcloud-calendar/
cp LICENSE LICENSE.upstream.md THIRD_PARTY_NOTICES.md dist/nextcloud-calendar/
mkdir -p dist/nextcloud-calendar/licenses
for notice in licenses/*; do install -m 644 "$notice" dist/nextcloud-calendar/licenses/; done
printf 'Corresponding source is included in the release archive under source/.\nProject: https://github.com/Ampleforth/remarkable2-nextcloudCalendar\n' > dist/nextcloud-calendar/SOURCE.txt
rsvg-convert -w 192 -h 192 qml/calendar.svg -o dist/nextcloud-calendar/icon.png
printf 'Built dist/nextcloud-calendar/ for reMarkable 2 (ARMv7).\n'
