#!/usr/bin/env bash
set -euo pipefail

# ============================================================
#  Pelican (Kaneil) Installer
#  One-stop installer for Panel (Go) + Wings (Go) daemon
#  github.com/YanIanZ/pelican-go
# ============================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
INSTALL_DIR="/opt/pelican"
PANEL_DIR="$PROJECT_DIR/panel"
WINGS_DIR="$PROJECT_DIR/wings"

# ---- Colors ----
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; CYAN='\033[0;36m'; NC='\033[0m'

# ---- Logging ----
log()   { echo -e "${BLUE}[KANEIL]${NC} $*"; }
ok()    { echo -e "${GREEN}[  OK  ]${NC} $*"; }
warn()  { echo -e "${YELLOW}[ WARN ]${NC} $*"; }
err()   { echo -e "${RED}[ ERROR]${NC} $*" >&2; }
banner() {
    echo ""
    echo -e "${CYAN}╔══════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║       ${GREEN}Pelican (Kaneil) Installer${CYAN}      ║${NC}"
    echo -e "${CYAN}║     Panel + Wings — All in Go         ║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════╝${NC}"
    echo ""
}

# ---- OS Detection ----
detect_os() {
    case "$(uname -s)" in
        Linux*)  OS="linux"
            if [ -f /etc/os-release ]; then
                . /etc/os-release
                DISTRO="$ID"
                VER="$VERSION_ID"
            elif [ -f /etc/debian_version ]; then
                DISTRO="debian"
                VER="$(cat /etc/debian_version)"
            elif [ -f /etc/redhat-release ]; then
                DISTRO="rhel"
            fi
            ;;
        Darwin*) OS="macos"; DISTRO="macos";;
        *) err "Unsupported OS: $(uname -s)"; exit 1;;
    esac
    log "Detected OS: $OS ($DISTRO $VER)"
}

# ---- Version Checks ----
check_version() {
    local cmd="$1" min="$2" name="$3"
    if command -v "$cmd" &>/dev/null; then
        local v; v=$("$cmd" --version 2>&1 | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1)
        if [ -n "$v" ]; then
            if printf '%s\n%s\n' "$min" "$v" | sort -V -C 2>/dev/null || [ "$(printf '%s\n' "$min" "$v" | sort -V | head -1)" = "$min" ]; then
                ok "$name $v (>= $min required)"
                return 0
            fi
        fi
    fi
    return 1
}

# ---- Dependency: Go ----
install_go() {
    local GO_VER="1.25.7"
    if check_version "go" "1.25" "Go"; then
        return 0
    fi

    warn "Go $GO_VER not found. Installing..."
    local arch; arch=$(uname -m)
    case "$arch" in
        x86_64)  arch="amd64";;
        aarch64) arch="arm64";;
        arm64)   arch="arm64";;
        *) err "Unsupported arch: $arch"; exit 1;;
    esac

    local go_tar="go${GO_VER}.${OS}-${arch}.tar.gz"
    local go_url="https://go.dev/dl/${go_tar}"

    log "Downloading Go $GO_VER for $OS/$arch..."
    cd /tmp
    curl -sSfL "$go_url" -o "$go_tar" || { err "Failed to download Go"; exit 1; }

    if [ -d /usr/local/go ]; then
        log "Removing old Go installation..."
        sudo rm -rf /usr/local/go
    fi

    sudo tar -C /usr/local -xzf "$go_tar"
    rm -f "$go_tar"

    # Add to PATH if not already there
    if ! grep -q '/usr/local/go/bin' ~/.profile 2>/dev/null; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
    fi
    export PATH="$PATH:/usr/local/go/bin"

    ok "Go $GO_VER installed"
}

# ---- Dependency: MySQL ----
install_mysql() {
    if command -v mysql &>/dev/null || command -v mariadb &>/dev/null; then
        ok "MySQL/MariaDB found"
        return 0
    fi

    warn "MySQL not found. Installing MariaDB..."
    case "$DISTRO" in
        ubuntu|debian)
            sudo apt-get update -qq
            sudo apt-get install -y -qq mariadb-server mariadb-client
            sudo systemctl enable --now mariadb
            ;;
        centos|rhel|fedora|rocky|almalinux)
            sudo dnf install -y mariadb-server mariadb
            sudo systemctl enable --now mariadb
            ;;
        macos)
            brew install mariadb
            brew services start mariadb
            ;;
    esac
    ok "MariaDB installed and running"
}

# ---- Dependency: Redis ----
install_redis() {
    if command -v redis-server &>/dev/null; then
        ok "Redis found"
        return 0
    fi

    warn "Redis not found. Installing..."
    case "$DISTRO" in
        ubuntu|debian)
            sudo apt-get update -qq
            sudo apt-get install -y -qq redis-server
            sudo systemctl enable --now redis-server
            ;;
        centos|rhel|fedora|rocky|almalinux)
            sudo dnf install -y redis
            sudo systemctl enable --now redis
            ;;
        macos)
            brew install redis
            brew services start redis
            ;;
    esac
    ok "Redis installed and running"
}

# ---- Dependency: Docker (for Wings) ----
install_docker() {
    if command -v docker &>/dev/null; then
        ok "Docker $(docker --version 2>/dev/null | awk '{print $3}' | tr -d ',') found"
        return 0
    fi

    warn "Docker not found. Installing..."
    case "$DISTRO" in
        ubuntu|debian)
            curl -fsSL https://get.docker.com | sudo bash
            sudo systemctl enable --now docker
            ;;
        centos|rhel|fedora|rocky|almalinux)
            curl -fsSL https://get.docker.com | sudo bash
            sudo systemctl enable --now docker
            ;;
        macos)
            err "Docker for macOS requires manual install. Visit: https://docs.docker.com/desktop/mac/install/"
            err "Install Docker Desktop and re-run this installer."
            exit 1
            ;;
    esac
    ok "Docker installed and running"

    # Add user to docker group
    if ! groups | grep -q docker; then
        sudo usermod -aG docker "$USER" 2>/dev/null || true
        warn "Added user to docker group. Re-login may be required."
    fi
}

# ---- Dependency: Git ----
install_git() {
    if command -v git &>/dev/null; then
        ok "Git found"
        return 0
    fi
    warn "Git not found. Installing..."
    case "$DISTRO" in
        ubuntu|debian)     sudo apt-get install -y -qq git;;
        centos|rhel|fedora|rocky|almalinux) sudo dnf install -y git;;
        macos)             brew install git;;
    esac
    ok "Git installed"
}

# ---- Install Panel ----
install_panel() {
    banner
    log "=== Installing Pelican Panel (Go) ==="
    echo ""

    install_go
    install_mysql
    install_redis

    # Build panel
    cd "$PANEL_DIR"
    if [ ! -f "go.mod" ]; then
        err "panel/go.mod not found. Is the repo cloned correctly?"
        exit 1
    fi

    log "Fetching Go dependencies..."
    go mod tidy 2>/dev/null || go mod download

    log "Building panel binary..."
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/panel" ./cmd/panel
    ok "Panel binary built: $INSTALL_DIR/panel"

    # Config
    if [ ! -f /etc/pelican/panel.yml ]; then
        log "Generating panel configuration..."
        sudo mkdir -p /etc/pelican
        sudo tee /etc/pelican/panel.yml > /dev/null << 'YAMLEOF'
app:
  name: Pelican
  env: production
  debug: false
  url: http://localhost
  timezone: UTC
  locale: en

database:
  driver: mysql
  host: 127.0.0.1
  port: 3306
  database: pelican
  username: pelican
  password: ""
  charset: utf8mb4
  collation: utf8mb4_unicode_ci

redis:
  host: 127.0.0.1
  port: 6379
  db: 0

auth:
  2fa_required: false
  2fa_bytes: 32
  2fa_window: 4
  verify_newer: true

api:
  key_limit: 25
  key_expire_time: 720

panel:
  use_binary_prefix: true
  editable_server_descriptions: true
  webhook_prune_days: 30
  files_max_edit_size: 4194304
YAMLEOF
        ok "Config: /etc/pelican/panel.yml"
        warn "Edit /etc/pelican/panel.yml to set your database password and app URL."
    else
        ok "Config already exists: /etc/pelican/panel.yml"
    fi

    # systemd service
    if [ "$OS" = "linux" ]; then
        log "Installing systemd service..."
        sudo tee /etc/systemd/system/pelican-panel.service > /dev/null << SERVICEOF
[Unit]
Description=Pelican Panel
After=network.target mysql.service mariadb.service redis.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/panel --config=/etc/pelican/panel.yml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
SERVICEOF
        sudo systemctl daemon-reload
        sudo systemctl enable pelican-panel
        sudo systemctl start pelican-panel 2>/dev/null || warn "Could not start panel (check config)"
        ok "Panel service installed"
    fi

    echo ""
    ok "=== Panel installation complete ==="
    echo "  Binary: $INSTALL_DIR/panel"
    echo "  Config: /etc/pelican/panel.yml"
    echo "  Start:  sudo systemctl start pelican-panel"
    echo "  Logs:   sudo journalctl -u pelican-panel -f"
}

# ---- Install Wings ----
install_wings() {
    banner
    log "=== Installing Pelican Wings (Go daemon) ==="
    echo ""

    install_go
    install_docker

    # Build wings
    cd "$WINGS_DIR"
    if [ ! -f "go.mod" ]; then
        err "wings/go.mod not found. Is the repo cloned correctly?"
        exit 1
    fi

    log "Fetching Go dependencies..."
    go mod tidy 2>/dev/null || go mod download

    log "Building wings binary..."
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/wings" .
    ok "Wings binary built: $INSTALL_DIR/wings"

    # Config
    if [ ! -f /etc/pelican/wings.yml ]; then
        log "Generating wings configuration..."
        sudo mkdir -p /etc/pelican
        sudo tee /etc/pelican/wings.yml > /dev/null << 'YAMLEOF'
debug: false
app_name: Pelican
uuid: ""
token_id: ""
token: ""
api:
    host: 0.0.0.0
    port: 8080
    ssl:
        enabled: false
        cert: /etc/letsencrypt/live/node.example.com/fullchain.pem
        key: /etc/letsencrypt/live/node.example.com/privkey.pem
    upload_limit: 100
system:
    data: /var/lib/pelican/volumes
    sftp:
        bind_port: 2022
        bind_address: 0.0.0.0
    log_directory: /var/log/pelican
    tmp_directory: /tmp/pelican
allowed_mounts: []
remote: https://your-panel.example.com
docker:
    network:
        name: pelican_nw
        interface: 172.18.0.1
        dns: [1.1.1.1, 8.8.8.8]
        driver: bridge
        is_internal: false
        enable_icc: true
        network_mode: pelican_nw
        network_mode_mc: pelican_nw
    domainname: ""
    registries: {}
    tmpfs_size: 100
    container_pid_limit: 512
    installer_limits:
        memory: 1024
        cpu: 100
    overhead:
        override: false
        default_multiplier: 1.05
        multipliers: {}
    use_performant_io_scheduler: true
throttles:
    enabled: true
    lines: 2000
    line_reset_interval: 100
crash_detection:
    enabled: true
    max_crashes: 3
    timeout: 60
backups:
    write_limit: 0
    compression_level: best_speed
transfers:
    download_limit: 0
pterodactyl:
    websocket_log_count: 150
YAMLEOF
        ok "Config: /etc/pelican/wings.yml"
        warn "Edit /etc/pelican/wings.yml to set uuid, token_id, token, and remote URL."
    else
        ok "Config already exists: /etc/pelican/wings.yml"
    fi

    # Data directories
    sudo mkdir -p /var/lib/pelican/volumes /var/log/pelican /tmp/pelican

    # systemd service
    if [ "$OS" = "linux" ]; then
        log "Installing systemd service..."
        sudo tee /etc/systemd/system/pelican-wings.service > /dev/null << SERVICEOF
[Unit]
Description=Pelican Wings Daemon
After=network.target docker.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/wings --config=/etc/pelican/wings.yml
Restart=always
RestartSec=10
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
SERVICEOF
        sudo systemctl daemon-reload
        sudo systemctl enable pelican-wings
        ok "Wings service installed"
        log "Start Wings after configuring /etc/pelican/wings.yml with node token from Panel:"
        log "  sudo systemctl start pelican-wings"
    fi

    echo ""
    ok "=== Wings installation complete ==="
    echo "  Binary: $INSTALL_DIR/wings"
    echo "  Config: /etc/pelican/wings.yml"
    echo "  Start:  sudo systemctl start pelican-wings"
    echo "  Logs:   sudo journalctl -u pelican-wings -f"
}

# ---- Full Install ----
install_all() {
    install_panel
    echo ""
    install_wings
    echo ""
    ok "=== Full installation complete ==="
    echo ""
    echo "  Next steps:"
    echo "  1. Edit /etc/pelican/panel.yml with your database password"
    echo "  2. Start: sudo systemctl start pelican-panel"
    echo "  3. Create a node in the Panel to get a token"
    echo "  4. Edit /etc/pelican/wings.yml with the node token"
    echo "  5. Start: sudo systemctl start pelican-wings"
}

# ---- Main Menu ----
main() {
    detect_os
    install_git

    banner
    echo "  Select installation:"
    echo "  1) Panel only (Go admin panel + API)"
    echo "  2) Wings only (Go daemon for game servers)"
    echo "  3) Full install (Panel + Wings)"
    echo "  4) Exit"
    echo ""
    read -rp "  Choice [1-4]: " choice

    case "$choice" in
        1) install_panel;;
        2) install_wings;;
        3) install_all;;
        4) log "Goodbye."; exit 0;;
        *) err "Invalid choice"; exit 1;;
    esac
}

main "$@"
