#!/bin/bash
# Quick bootstrap — curl this to install KaNeil (Pelican)
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/YanIanZ/pelican-go/main/kaneil-installer/bootstrap.sh | bash
#   curl -sSL https://raw.githubusercontent.com/YanIanZ/pelican-go/main/kaneil-installer/bootstrap.sh | PANEL_TAG=main bash -s -- --panel-only
#
# Or locally:
#   cd kaneil-installer && sudo ./install.sh
#

set -e

REPO="https://github.com/YanIanZ/pelican-go.git"
BRANCH="${BRANCH:-main}"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "==> Fetching KaNeil installer ($BRANCH)..."
cd "$TMPDIR"
git clone --depth 1 --branch "$BRANCH" "$REPO" pelican 2>/dev/null || {
    echo "==> Git clone failed, downloading raw scripts..."
    mkdir -p pelican/kaneil-installer
    curl -sSL "https://raw.githubusercontent.com/YanIanZ/pelican-go/${BRANCH}/kaneil-installer/install.sh" -o pelican/kaneil-installer/install.sh
    chmod +x pelican/kaneil-installer/install.sh
}

echo "==> Running installer..."
cd pelican/kaneil-installer
exec sudo bash install.sh "$@"
