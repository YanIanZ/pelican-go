#!/bin/bash
# KaNeil (Pelican) — Bootstrap Installer
#
# Clone and run (one command):
#   git clone https://github.com/YanIanZ/pelican-go.git && cd pelican-go/kaneil-installer && sudo ./install.sh
#
set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
exec sudo bash "$SCRIPT_DIR/install.sh" "$@"