#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
#  Pelican (Go) Installer — Panel + Wings, both in Go
#  github.com/YanIanZ/pelican-go
#
#  Usage:
#    bash <(curl -fsSL https://raw.githubusercontent.com/YanIanZ/pelican-go/main/install.sh)
#
#  Options (env vars):
#    FQDN=my.host MYSQL_PASSWORD=secret INSTALL_WINGS=no sudo -E bash install.sh
# =============================================================================

[[ $EUID -eq 0 ]] || { echo "[!] must run as root" >&2; exit 1; }
export DEBIAN_FRONTEND=noninteractive

# ─── config (override via env) ──────────────────────────────────────────────

INSTALL_PANEL="${INSTALL_PANEL:-yes}"
INSTALL_WINGS="${INSTALL_WINGS:-yes}"
FQDN="${FQDN:-$(hostname -f 2>/dev/null || echo localhost)}"
MYSQL_DB="${MYSQL_DB:-pelican}"
MYSQL_USER="${MYSQL_USER:-pelican}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-$(head -c24 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c24)}"
APP_SECRET="${APP_SECRET:-$(head -c32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c32)}"
GO_VERSION="${GO_VERSION:-1.25.7}"
INSTALL_DIR="/opt/pelican"
CONFIG_DIR="/etc/pelican"
FIREWALL="${CONFIGURE_FIREWALL:-no}"
SSL="${ASSUME_SSL:-no}"
REPO_URL="https://github.com/YanIanZ/pelican-go.git"
REPO_BRANCH="${REPO_BRANCH:-main}"
PANEL_SRC=""
WINGS_SRC=""

# ─── helpers ─────────────────────────────────────────────────────────────────

R='\033[0;31m'; G='\033[0;32m'; Y='\033[1;33m'; B='\033[0;34m'; N='\033[0m'
log()  { echo -e " ${B}→${N} $*"; }
ok()   { echo -e " ${G}✓${N} $*"; }
warn() { echo -e " ${Y}!${N} $*"; }
die()  { echo -e "${R}✗ $*${N}" >&2; exit 1; }

check_cmd() { command -v "$1" &>/dev/null; }
finish()   { rm -rf /tmp/pelican-src; }

banner() {
    echo ""
    echo -e "${B}╔══════════════════════════════════════════════╗${N}"
    echo -e "${B}║     ${G}Pelican (Go) — Panel + Wings Installer${B}     ║${N}"
    echo -e "${B}╚══════════════════════════════════════════════╝${N}"
    echo ""
}

# ─── OS / arch detection ────────────────────────────────────────────────────

detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS_ID="$ID"
        OS_VER="${VERSION_ID%%.*}"
    elif [ -f /etc/debian_version ]; then
        OS_ID="debian"
        OS_VER="$(cat /etc/debian_version | cut -d. -f1)"
    else
        die "unsupported OS"
    fi

    case "$OS_ID" in
        ubuntu|debian)            PKG="apt-get install -y -qq";;
        rocky|almalinux|rhel)    PKG="dnf install -y -q";;
        fedora)                  PKG="dnf install -y -q";;
        *) die "unsupported OS: $OS_ID";;
    esac

    case "$(uname -m)" in
        x86_64)  GO_ARCH="amd64"  ;;
        arm64|aarch64) GO_ARCH="arm64" ;;
        *) die "unsupported arch: $(uname -m)";;
    esac

    log "detected: $OS_ID $OS_VER / $GO_ARCH"
}

pkg()   { $PKG "$@" 2>/dev/null || $PKG "$@"; }
update(){ [ "$OS_ID" = "debian" ] || [ "$OS_ID" = "ubuntu" ] && apt-get update -qq; }

# ─── dependency installers ───────────────────────────────────────────────────

install_go() {
    local want="1.25"
    if check_cmd go; then
        local have; have=$(go version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -1)
        if printf '%s\n%s\n' "$want" "$have" | sort -V -C 2>/dev/null; then
            ok "go $have"
            return
        fi
    fi
    log "installing go $GO_VERSION..."
    local tgz="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    curl -fsSL "https://go.dev/dl/${tgz}" -o "/tmp/${tgz}"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "/tmp/${tgz}"
    rm -f "/tmp/${tgz}"
    export PATH="/usr/local/go/bin:$PATH"
    grep -q '/usr/local/go/bin' /etc/profile 2>/dev/null || echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    ok "go $GO_VERSION"
}

install_docker() {
    check_cmd docker && { ok "docker"; return; }
    log "installing docker..."
    curl -fsSL https://get.docker.com | bash
    systemctl enable --now docker
    ok "docker"
}

install_mariadb() {
    check_cmd mariadb || check_cmd mysql && { ok "mariadb/mysql"; return; }
    log "installing mariadb..."
    update
    case "$OS_ID" in
        ubuntu|debian) pkg mariadb-server mariadb-client;;
        *) curl -fsSL https://downloads.mariadb.com/MariaDB/mariadb_repo_setup | bash; pkg MariaDB-server MariaDB-client;;
    esac
    systemctl enable --now mariadb 2>/dev/null || systemctl enable --now mysql 2>/dev/null || true
    ok "mariadb"
}

install_redis() {
    check_cmd redis-server && { ok "redis"; return; }
    log "installing redis..."
    update; pkg redis-server 2>/dev/null || pkg redis
    systemctl enable --now redis-server 2>/dev/null || systemctl enable --now redis 2>/dev/null || true
    ok "redis"
}

install_nginx() {
    check_cmd nginx && { ok "nginx"; return; }
    log "installing nginx..."
    update; pkg nginx
    systemctl enable --now nginx
    ok "nginx"
}

# ─── source resolution ───────────────────────────────────────────────────────

resolve_source() {
    local sdir; sdir="$(cd "$(dirname "$0")" && pwd)"
    local pdir; pdir="$(dirname "$sdir")"

    # check if sibling panel/wings exist (local repo)
    if [ -f "$pdir/panel/go.mod" ] && [ -f "$pdir/wings/go.mod" ]; then
        PANEL_SRC="$pdir/panel"
        WINGS_SRC="$pdir/wings"
        ok "using local source"
        return
    fi

    # check env override
    if [ -n "${PANEL_SRC:-}" ] && [ -n "${WINGS_SRC:-}" ]; then
        [ -f "$PANEL_SRC/go.mod" ] || die "PANEL_SRC not found: $PANEL_SRC"
        [ -f "$WINGS_SRC/go.mod" ] || die "WINGS_SRC not found: $WINGS_SRC"
        return
    fi

    # clone from GitHub
    log "cloning pelican-go ($REPO_BRANCH)..."
    rm -rf /tmp/pelican-src
    git clone --depth 1 --branch "$REPO_BRANCH" "$REPO_URL" /tmp/pelican-src
    PANEL_SRC="/tmp/pelican-src/panel"
    WINGS_SRC="/tmp/pelican-src/wings"
    ok "source from GitHub"
}

# ─── build ───────────────────────────────────────────────────────────────────

build_panel() {
    log "building panel (downloading Go modules, this may take a minute)..."
    cd "$PANEL_SRC"
    export GOPROXY="${GOPROXY:-https://proxy.golang.org,direct}"
    go mod download -x 2>/dev/null &
    local pid=$!
    local dots="."
    while kill -0 $pid 2>/dev/null; do
        printf "\r → downloading modules%s   " "$dots"
        dots="${dots}."
        [ ${#dots} -gt 30 ] && dots="."
        sleep 2
    done
    wait $pid || true
    echo ""
    log "compiling panel..."
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/panel" ./cmd/panel
    ok "panel → $INSTALL_DIR/panel"
}

build_wings() {
    log "building wings (downloading Go modules, this may take a minute)..."
    cd "$WINGS_SRC"
    export GOPROXY="${GOPROXY:-https://proxy.golang.org,direct}"
    go mod download -x 2>/dev/null &
    local pid=$!
    local dots="."
    while kill -0 $pid 2>/dev/null; do
        printf "\r → downloading modules%s   " "$dots"
        dots="${dots}."
        [ ${#dots} -gt 30 ] && dots="."
        sleep 2
    done
    wait $pid || true
    echo ""
    log "compiling wings..."
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/wings" .
    ok "wings → $INSTALL_DIR/wings"
}

# ─── config files ────────────────────────────────────────────────────────────

write_panel_conf() {
    log "writing $CONFIG_DIR/panel.yml..."
    mkdir -p "$CONFIG_DIR"
    cat > "$CONFIG_DIR/panel.yml" << EOF
app:
  name: Pelican
  env: production
  debug: false
  url: "${PROTO:-http}://${FQDN}"
  timezone: UTC
  locale: en
  secret: "${APP_SECRET}"

database:
  driver: mysql
  host: 127.0.0.1
  port: 3306
  database: ${MYSQL_DB}
  username: ${MYSQL_USER}
  password: "${MYSQL_PASSWORD}"
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
EOF
    ok "panel config"
}

write_wings_conf() {
    log "writing $CONFIG_DIR/config.yml..."
    mkdir -p "$CONFIG_DIR"
    mkdir -p /var/lib/pelican/{volumes,archives,backups} /var/log/pelican /tmp/pelican

    local ssl=false; [ "$SSL" = "yes" ] && ssl=true
    local fl; fl=$(echo "$FQDN" | tr '[:upper:]' '[:lower:]')

    cat > "$CONFIG_DIR/config.yml" << EOF
debug: false
app_name: Pelican
uuid: ""
token_id: ""
token: ""

api:
  host: 0.0.0.0
  port: 8080
  ssl:
    enabled: ${ssl}
    cert: /etc/letsencrypt/live/${fl}/fullchain.pem
    key: /etc/letsencrypt/live/${fl}/privkey.pem
  upload_limit: 100

system:
  root_directory: /var/lib/pelican
  log_directory: /var/log/pelican
  data: /var/lib/pelican/volumes
  archive_directory: /var/lib/pelican/archives
  backup_directory: /var/lib/pelican/backups
  tmp_directory: /tmp/pelican
  username: pelican
  sftp:
    bind_address: 0.0.0.0
    bind_port: 2022

remote: "${PROTO:-http}://${FQDN}"

allowed_mounts: []

docker:
  network:
    name: pelican_nw
    interface: 172.18.0.1
    dns:
      - 1.1.1.1
      - 8.8.8.8
    driver: bridge
    is_internal: false
    enable_icc: true
    network_mode: pelican_nw
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
  detect_clean_exit_as_crash: true
  timeout: 60

backups:
  write_limit: 0
  compression_level: best_speed

transfers:
  download_limit: 0

allowed_origins:
  - "${PROTO:-http}://${FQDN}"
EOF
    ok "wings config"
}

# ─── database ────────────────────────────────────────────────────────────────

setup_db() {
    log "setting up database..."
    local m="mariadb"; check_cmd mysql && m="mysql"
    $m -u root << SQL 2>/dev/null || true
CREATE DATABASE IF NOT EXISTS \`${MYSQL_DB}\`;
CREATE USER IF NOT EXISTS '${MYSQL_USER}'@'127.0.0.1' IDENTIFIED BY '${MYSQL_PASSWORD}';
ALTER USER '${MYSQL_USER}'@'127.0.0.1' IDENTIFIED BY '${MYSQL_PASSWORD}';
GRANT ALL PRIVILEGES ON \`${MYSQL_DB}\`.* TO '${MYSQL_USER}'@'127.0.0.1' WITH GRANT OPTION;
FLUSH PRIVILEGES;
SQL
    ok "database $MYSQL_DB"
}

# ─── nginx ───────────────────────────────────────────────────────────────────

setup_nginx() {
    log "setting up nginx for $FQDN..."
    local conf="/etc/nginx/sites-available/pelican"
    mkdir -p /etc/nginx/sites-{available,enabled}
    cat > "$conf" << NGINX
server {
    listen 80;
    server_name ${FQDN};

    client_max_body_size 100m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /api/client/servers/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }
}
NGINX
    ln -sf "$conf" /etc/nginx/sites-enabled/pelican 2>/dev/null
    rm -f /etc/nginx/sites-enabled/default
    nginx -t && systemctl reload nginx
    ok "nginx → $FQDN"
}

# ─── systemd services ────────────────────────────────────────────────────────

install_services() {
    mkdir -p "$INSTALL_DIR"

    if [ "$INSTALL_PANEL" = "yes" ]; then
        log "installing pelican-panel service..."
        cat > /etc/systemd/system/pelican-panel.service << UNIT
[Unit]
Description=Pelican Panel
After=network.target mariadb.service mysql.service redis.service redis-server.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/panel --config=$CONFIG_DIR/panel.yml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
UNIT
        systemctl daemon-reload
        systemctl enable pelican-panel
        systemctl start pelican-panel 2>/dev/null || warn "panel failed to start — check journalctl -u pelican-panel"
        ok "pelican-panel service"
    fi

    if [ "$INSTALL_WINGS" = "yes" ]; then
        log "installing pelican-wings service..."
        cat > /etc/systemd/system/pelican-wings.service << UNIT
[Unit]
Description=Pelican Wings Daemon
After=network.target docker.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/wings --config=$CONFIG_DIR/config.yml
Restart=always
RestartSec=10
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
UNIT
        systemctl daemon-reload
        systemctl enable pelican-wings
        ok "pelican-wings service (stopped — set token first)"
    fi
}

# ─── firewall ────────────────────────────────────────────────────────────────

setup_firewall() {
    [ "$FIREWALL" != "yes" ] && return
    log "configuring firewall..."
    if check_cmd ufw; then
        for p in 80 443 2022 8080; do ufw allow "$p"/tcp 2>/dev/null; done
        ufw --force reload
    elif check_cmd firewall-cmd; then
        for p in 80 443 2022 8080; do firewall-cmd --zone=public --add-port="$p"/tcp --permanent 2>/dev/null; done
        firewall-cmd --reload
    fi
    ok "firewall"
}

# ─── summary ─────────────────────────────────────────────────────────────────

summary() {
    echo ""
    echo -e "${G}══════════════════════════════════════════════════════${N}"
    echo -e "${G}               Installation Complete                   ${N}"
    echo -e "${G}══════════════════════════════════════════════════════${N}"
    echo ""

    if [ "$INSTALL_PANEL" = "yes" ]; then
        echo -e "  ${B}Panel${N}"
        echo "  ├─ URL:      ${PROTO:-http}://${FQDN}"
        echo "  ├─ Binary:   $INSTALL_DIR/panel"
        echo "  ├─ Config:   $CONFIG_DIR/panel.yml"
        echo "  ├─ Service:  systemctl status pelican-panel"
        echo "  ├─ Logs:     journalctl -u pelican-panel -f"
        echo "  ├─ DB name:  $MYSQL_DB"
        echo "  ├─ DB user:  $MYSQL_USER"
        echo "  └─ DB pass:  $MYSQL_PASSWORD"
        echo ""
    fi

    if [ "$INSTALL_WINGS" = "yes" ]; then
        echo -e "  ${B}Wings${N}"
        echo "  ├─ Binary:   $INSTALL_DIR/wings"
        echo "  ├─ Config:   $CONFIG_DIR/config.yml"
        echo "  └─ Service:  systemctl start pelican-wings"
        echo "     (set uuid, token_id, token in config.yml first!)"
        echo ""
    fi

    echo -e "  ${Y}Next:${N}"
    if [ "$INSTALL_PANEL" = "yes" ]; then
        echo "  1. Open ${PROTO:-http}://${FQDN}"
        echo "  2. Create an admin user"
    fi
    if [ "$INSTALL_WINGS" = "yes" ]; then
        echo "  3. Add a Node in the Panel → copy uuid, token_id, token"
        echo "  4. Paste into $CONFIG_DIR/config.yml"
        echo "  5. systemctl start pelican-wings"
    fi
    echo ""
}

# ─── main ────────────────────────────────────────────────────────────────────

trap finish EXIT

banner
detect_os

log "installing base packages..."
update
pkg curl git certbot

log "resolving source..."
resolve_source

if [ "$INSTALL_PANEL" = "yes" ]; then
    install_go
    install_mariadb
    install_redis
    install_nginx
    build_panel
    setup_db
    write_panel_conf
fi

if [ "$INSTALL_WINGS" = "yes" ]; then
    install_go
    install_docker
    build_wings
    write_wings_conf
fi

install_services
[ "$INSTALL_PANEL" = "yes" ] && setup_nginx
setup_firewall
summary
