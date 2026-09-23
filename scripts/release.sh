#!/bin/sh
# release.sh — bump version, commit, tag. Usage: ./scripts/release.sh 0.2.0
set -e
if [ -z "$1" ]; then
  echo "uso: $0 <version sin v>" >&2
  exit 2
fi
VER="$1"
sed -i "s/\"version\": \".*\"/\"version\": \"$VER\"/" package.json
grep '"version"' package.json
git add package.json
git commit -m "chore: release v$VER" || true
git tag "v$VER"
echo "tag v$VER listo. Pusheá con: git push origin main v$VER"
