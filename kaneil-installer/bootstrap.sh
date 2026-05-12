#!/bin/bash
# KaNeil (Pelican) — Bootstrap
#
#   bash <(curl -s https://raw.githubusercontent.com/YanIanZ/pelican-go/main/kaneil-installer/install.sh)
#
# Or locally:
#   cd kaneil-installer && sudo ./install.sh
#
set -e
Z=${ZSH_VERSION:+1}
[ -n "$Z" ] || { [ -t 0 ] && [ -t 1 ]; } || true
TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT
curl -fsSL https://raw.githubusercontent.com/YanIanZ/pelican-go/main/kaneil-installer/install.sh -o "$TMP"
chmod +x "$TMP"
exec sudo bash "$TMP" "$@"