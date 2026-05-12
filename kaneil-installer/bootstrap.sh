#!/bin/bash
# KaNeil (Pelican) — Quick Bootstrap
# 
# One-command install:
#   curl -fsSL raw.githubusercontent.com/YanIanZ/pelican-go/main/kaneil-installer/bootstrap.sh | sudo bash
#
# Options (semicolon-separated env vars):
#   curl ... | INSTALL_WINGS=no sudo -E bash   # panel only
#   curl ... | FQDN=my.host MYSQL_PASSWORD=xxx sudo -E bash
#
set -e
BRANCH="${BRANCH:-main}"
BASE="https://raw.githubusercontent.com/YanIanZ/pelican-go/${BRANCH}/kaneil-installer"
TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT
echo "=> fetching installer (${BRANCH})..."
curl -fsSL "${BASE}/install.sh" -o "$TMP"
chmod +x "$TMP"
exec bash "$TMP" "$@"