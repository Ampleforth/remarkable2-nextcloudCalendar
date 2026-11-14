#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-only
set -euo pipefail
cd -- "$(dirname -- "$0")/.."
version=${1:-$(cat VERSION)}
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'Use a numeric X.Y.Z version' >&2; exit 2; }
target_arch=${2:-arm}
case "$target_arch" in
  arm) app=dist/nextcloud-calendar; suffix=rm1-rm2-armv7 ;;
  arm64) app=dist/nextcloud-calendar-arm64; suffix=paper-pro-aarch64 ;;
  *) echo 'Second argument must be arm or arm64.' >&2; exit 2 ;;
esac
[[ -x "$app/backend/entry" && -s "$app/resources.rcc" && -s "$app/icon.png" ]] || {
  echo 'Build the app with ./scripts/build.sh first.' >&2; exit 1;
}
[[ -s docs/images/week.png && -s docs/images/tasks.png ]] || {
  echo 'Generate the example images with ./scripts/screenshots.sh first.' >&2; exit 1;
}
python3 scripts/audit.py .
mkdir -p release
stage=$(mktemp -d)
trap 'rm -rf -- "$stage"' EXIT
name="remarkable2-nextcloudCalendar-v${version}-${suffix}"
root="$stage/$name"
mkdir -p "$root/source"
cp -R "$app" "$root/nextcloud-calendar"
cp scripts/install.sh "$root/install.sh"
chmod 755 "$root/install.sh"
cp VERSION CHANGELOG.md README.md INSTALL.md SECURITY.md LICENSE LICENSE.upstream.md THIRD_PARTY_NOTICES.md config.example.json "$root/"
cp -R docs licenses "$root/"
# Include the exact corresponding source and vendored dependency source.
tar --exclude='__pycache__' --exclude='*.pyc' -cf - \
  *.go go.mod go.sum VERSION CHANGELOG.md README.md INSTALL.md SECURITY.md LICENSE LICENSE.upstream.md \
  THIRD_PARTY_NOTICES.md config.example.json .gitignore .gitattributes \
  qml packaging scripts tests examples vendor docs licenses | tar -xf - -C "$root/source"

python3 scripts/audit.py "$root"
tar -czf "release/$name.tar.gz" -C "$stage" "$name"
(cd release && sha256sum "$name.tar.gz" > "$name.sha256"
  cat *.sha256 > SHA256SUMS)
printf 'Packaged release/%s.tar.gz and release/SHA256SUMS\n' "$name"
