#!/usr/bin/env bash
# shellcheck disable=SC2034,SC2155
set -euo pipefail

#######################################################################################
#                                                                                     #
#   Pelican (KaNeil) — Go Panel + Wings Installer                                     #
#                                                                                     #
#   github.com/YanIanZ/pelican-go                                                     #
#                                                                                     #
#   One-command install from local source.                                       #
#                                                                                     #
#   Usage (from repo root):                                                           #
#     cd kaneil-installer && sudo ./install.sh                                        #
#                                                                                     #
#   Env vars (non-interactive):                                                       #
#     FQDN=MYSQL_PASSWORD=... sudo -E bash install.sh                                 #
#                                                                                     #
#######################################################################################

[[ $EUID -ne 0 ]] && { echo "ERROR: must run as root" >&2; exit 1; }

# ── configuration (override via env) ────────────────────────────────────────────────
export DEBIAN_FRONTEND=noninteractive
GO_VERSION="${GO_VERSION:-1.25.7}"
INSTALL_DIR="${INSTALL_DIR:-/opt/pelican}"
CONFIG_DIR="${CONFIG_DIR:-/etc/pelican}"

# what to install
INSTALL_PANEL="${INSTALL_PANEL:-yes}"
INSTALL_WINGS="${INSTALL_WINGS:-yes}"
INSTALL_NGINX="${INSTALL_NGINX:-yes}"
CONFIGURE_FIREWALL="${CONFIGURE_FIREWALL:-no}"
ASSUME_SSL="${ASSUME_SSL:-no}"

# panel config
PANEL_FQDN="${FQDN:-$(hostname -f 2>/dev/null || echo localhost)}"
PANEL_EMAIL="${EMAIL:-}"
MYSQL_DB="${MYSQL_DB:-pelican}"
MYSQL_USER="${MYSQL_USER:-pelican}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-$(head -c24 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c24)}"
APP_SECRET="${APP_SECRET:-$(head -c32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c32)}"

# wings config
WINGS_UUID="${WINGS_UUID:-}"
WINGS_TOKEN_ID="${WINGS_TOKEN_ID:-}"
WINGS_TOKEN="${WINGS_TOKEN:-}"

# ── helpers ─────────────────────────────────────────────────────────────────────────
R='\033[0;31m'; G='\033[0;32m'; Y='\033[1;33m'; B='\033[0;34m'; N='\033[0m'
OK="${G}✓${N}"; WARN="${Y}!${N}"

say()    { echo -e " ${B}→${N} $*"; }
ok()     { echo -e " ${OK} $*"; }
alert()  { echo -e " ${WARN} $*"; }
die()    { echo -e "${R}✗ $*${N}" >&2; exit 1; }

header() {
    echo ""
    echo -e "${B}╔══════════════════════════════════════════════════╗${N}"
    echo -e "${B}║      ${G}Pelican (KaNeil) — Go Full-Stack Installer${B}      ║${N}"
    echo -e "${B}║          Panel + Wings built from source          ║${N}"
    echo -e "${B}╚══════════════════════════════════════════════════╝${N}"
    echo ""
}

check_cmd() { command -v "$1" &>/dev/null; }

os_detect() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS_ID="${ID}"; OS_VER="${VERSION_ID%%.*}"
    elif [ -f /etc/debian_version ]; then
        OS_ID="debian"; OS_VER="$(cat /etc/debian_version | cut -d. -f1)"
    else
        die "unsupported OS — Ubuntu / Debian / Rocky / AlmaLinux only"
    fi

    case "$(uname -m)" in
        x86_64)  GO_ARCH="amd64"  ;;
        arm64|aarch64) GO_ARCH="arm64" ;;
        *) die "unsupported architecture: $(uname -m)" ;;
    esac
    say "detected: $OS_ID $OS_VER / $GO_ARCH"
}

# ── package helpers ─────────────────────────────────────────────────────────────────
pkg() {
    case "$OS_ID" in
        ubuntu|debian)          apt-get install -y -qq "$@" ;;
        rocky|almalinux|rhel)   dnf install -y -q "$@" ;;
        fedora)                 dnf install -y -q "$@" ;;
    esac
}

pkg_update() {
    case "$OS_ID" in
        ubuntu|debian) apt-get update -qq ;;
    esac
}

# ── dependency installers ───────────────────────────────────────────────────────────

need_go() {
    local want="1.25"
    if check_cmd go; then
        local have; have=$(go version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -1)
        if printf '%s\n%s\n' "$want" "$have" | sort -V -C 2>/dev/null; then
            ok "go $have"
            return
        fi
    fi
    say "installing go ${GO_VERSION}..."
    local tgz="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    curl -fsSL "https://go.dev/dl/${tgz}" -o "/tmp/${tgz}"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "/tmp/${tgz}"
    rm -f "/tmp/${tgz}"
    export PATH="/usr/local/go/bin:$PATH"
    grep -qx 'export PATH=.*/usr/local/go/bin' /etc/profile 2>/dev/null || \
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    ok "go ${GO_VERSION}"
}

need_docker() {
    if check_cmd docker; then ok "docker"; return; fi
    say "installing docker..."
    curl -fsSL https://get.docker.com | bash
    systemctl enable --now docker
    ok "docker"
}

need_mariadb() {
    if check_cmd mariadb || check_cmd mysql; then ok "mariadb/mysql"; return; fi
    say "installing mariadb..."
    pkg_update
    case "$OS_ID" in
        ubuntu|debian) pkg mariadb-server mariadb-client ;;
        rocky|almalinux|rhel|fedora)
            curl -fsSL https://downloads.mariadb.com/MariaDB/mariadb_repo_setup | bash
            pkg MariaDB-server MariaDB-client ;;
    esac
    systemctl enable --now mariadb 2>/dev/null || systemctl enable --now mysql 2>/dev/null || true
    ok "mariadb"
}

need_redis() {
    if check_cmd redis-server; then ok "redis"; return; fi
    say "installing redis..."
    pkg_update
    pkg redis-server 2>/dev/null || pkg redis
    systemctl enable --now redis-server 2>/dev/null || systemctl enable --now redis 2>/dev/null || true
    ok "redis"
}

need_nginx() {
    if check_cmd nginx; then ok "nginx"; return; fi
    say "installing nginx..."
    pkg_update
    pkg nginx
    systemctl enable --now nginx
    ok "nginx"
}

need_curl()     { check_cmd curl     || { pkg_update; pkg curl;     }; ok "curl"; }
need_certbot()  { check_cmd certbot  || { pkg_update; pkg certbot;  }; ok "certbot"; }

# ── build from local source ──────────────────────────────────────────────────────────
# installer lives in kaneil-installer/, panel/ and wings/ are siblings
_script_dir="$(cd "$(dirname "$0")" && pwd)"
_project_dir="$(dirname "$_script_dir")"

build_panel() {
    local src="${PANEL_SRC:-$_project_dir/panel}"
    [ -f "$src/go.mod" ] || die "panel source not found at $src"
    say "building panel..."
    cd "$src"
    go mod download
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/panel" ./cmd/panel
    ok "panel → $INSTALL_DIR/panel"
}

build_wings() {
    local src="${WINGS_SRC:-$_project_dir/wings}"
    [ -f "$src/go.mod" ] || die "wings source not found at $src"
    say "building wings..."
    cd "$src"
    go mod download
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/wings" .
    ok "wings → $INSTALL_DIR/wings"
}

# ── config generators ───────────────────────────────────────────────────────────────

write_panel_config() {
    say "writing $CONFIG_DIR/panel.yml..."
    mkdir -p "$CONFIG_DIR"
    cat > "$CONFIG_DIR/panel.yml" << EOF
app:
  name: Pelican
  env: production
  debug: false
  url: "${PROTO:-http}://${PANEL_FQDN}"
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

write_wings_config() {
    say "writing $CONFIG_DIR/config.yml..."
    mkdir -p "$CONFIG_DIR" /var/lib/pelican/volumes /var/lib/pelican/archives /var/lib/pelican/backups /var/log/pelican /tmp/pelican

    local fqdn_lower; fqdn_lower=$(echo "$PANEL_FQDN" | tr '[:upper:]' '[:lower:]')
    local ssl_enabled="false"; [ "$ASSUME_SSL" = "yes" ] && ssl_enabled="true"

    cat > "$CONFIG_DIR/config.yml" << EOF
debug: false
app_name: Pelican
uuid: "${WINGS_UUID}"
token_id: "${WINGS_TOKEN_ID}"
token: "${WINGS_TOKEN}"

api:
  host: 0.0.0.0
  port: 8080
  ssl:
    enabled: ${ssl_enabled}
    cert: /etc/letsencrypt/live/${fqdn_lower}/fullchain.pem
    key: /etc/letsencrypt/live/${fqdn_lower}/privkey.pem
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

remote: "${PROTO:-http}://${PANEL_FQDN}"

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
  - "${PROTO:-http}://${PANEL_FQDN}"
EOF
    ok "wings → $CONFIG_DIR/config.yml"
}

# ── database ────────────────────────────────────────────────────────────────────────

setup_db() {
    say "setting up database..."
    local mysql="mariadb"
    check_cmd mysql && mysql="mysql"

    $mysql -u root << SQL 2>/dev/null || true
CREATE DATABASE IF NOT EXISTS \`${MYSQL_DB}\`;
CREATE USER IF NOT EXISTS '${MYSQL_USER}'@'127.0.0.1' IDENTIFIED BY '${MYSQL_PASSWORD}';
ALTER USER '${MYSQL_USER}'@'127.0.0.1' IDENTIFIED BY '${MYSQL_PASSWORD}';
GRANT ALL PRIVILEGES ON \`${MYSQL_DB}\`.* TO '${MYSQL_USER}'@'127.0.0.1' WITH GRANT OPTION;
FLUSH PRIVILEGES;
SQL
    ok "database ${MYSQL_DB}"
}

# ── nginx ───────────────────────────────────────────────────────────────────────────

setup_nginx() {
    say "setting up nginx for ${PANEL_FQDN}..."
    local conf="/etc/nginx/sites-available/pelican"

    mkdir -p /etc/nginx/sites-available /etc/nginx/sites-enabled

    cat > "$conf" << NGINX
server {
    listen 80;
    server_name ${PANEL_FQDN};

    client_max_body_size 100m;
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml;

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
    ok "nginx → $PANEL_FQDN"
}

# ── systemd ─────────────────────────────────────────────────────────────────────────

install_services() {
    mkdir -p "$INSTALL_DIR"

    if [ "$INSTALL_PANEL" = "yes" ]; then
        say "installing pelican-panel service..."
        cat > /etc/systemd/system/pelican-panel.service << UNIT
[Unit]
Description=Pelican Panel
After=network.target mariadb.service mysql.service redis.service redis-server.service
Wants=mariadb.service redis.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/panel --config=$CONFIG_DIR/panel.yml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
UNIT
        systemctl daemon-reload
        systemctl enable pelican-panel
        systemctl start pelican-panel 2>/dev/null \
            || alert "panel didn't start — check journalctl -u pelican-panel"
        ok "pelican-panel service"
    fi

    if [ "$INSTALL_WINGS" = "yes" ]; then
        say "installing pelican-wings service..."
        cat > /etc/systemd/system/pelican-wings.service << UNIT
[Unit]
Description=Pelican Wings Daemon
After=network.target docker.service
Wants=docker.service

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
        ok "pelican-wings service (stopped — configure token first!)"
    fi
}

# ── firewall ────────────────────────────────────────────────────────────────────────

setup_firewall() {
    [ "$CONFIGURE_FIREWALL" != "yes" ] && return
    say "configuring firewall..."

    if check_cmd ufw; then
        for p in 80 443 2022 8080; do ufw allow "$p"/tcp 2>/dev/null; done
        ufw --force reload
    elif check_cmd firewall-cmd; then
        for p in 80 443 2022 8080; do firewall-cmd --zone=public --add-port="$p"/tcp --permanent 2>/dev/null; done
        firewall-cmd --reload
    fi
    ok "firewall"
}

# ── show summary ────────────────────────────────────────────────────────────────────

summary() {
    echo ""
    echo -e "${G}══════════════════════════════════════════════════════${N}"
    echo -e "${G}               Installation Complete                   ${N}"
    echo -e "${G}══════════════════════════════════════════════════════${N}"
    echo ""

    if [ "$INSTALL_PANEL" = "yes" ]; then
        echo -e "  ${B}Panel${N}"
        echo "  ├─ URL:      ${PROTO:-http}://${PANEL_FQDN}"
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

    if [ "$INSTALL_WINGS" = "yes" ]; then
        echo -e "  ${Y}Next steps:${N}"
        echo "  1. Open ${PROTO:-http}://${PANEL_FQDN} in your browser"
        echo "  2. Create an admin user via the panel"
        echo "  3. Add a Node in the admin panel to generate a Wings token"
        echo "  4. Copy uuid + token_id + token into $CONFIG_DIR/config.yml"
        echo "  5. Start Wings: systemctl start pelican-wings"
    else
        echo -e "  ${Y}Next steps:${N}"
        echo "  1. Open ${PROTO:-http}://${PANEL_FQDN} in your browser"
        echo "  2. Create an admin user"
    fi
    echo ""
}

# ── main ────────────────────────────────────────────────────────────────────────────

main() {
    header
    os_detect

    say "checking dependencies..."
    need_curl

    if [ "$INSTALL_PANEL" = "yes" ]; then
        need_mariadb
        need_redis
        need_nginx
    fi

    if [ "$INSTALL_WINGS" = "yes" ]; then
        need_docker
    fi

    need_go
    need_certbot
    say "all dependencies satisfied"

    mkdir -p "$INSTALL_DIR"

    if [ "$INSTALL_PANEL" = "yes" ]; then
        build_panel
        setup_db
        write_panel_config
    fi

    if [ "$INSTALL_WINGS" = "yes" ]; then
        build_wings
        write_wings_config
    fi

    install_services

    if [ "$INSTALL_NGINX" = "yes" ] && [ "$INSTALL_PANEL" = "yes" ]; then
        setup_nginx
    fi

    setup_firewall
    summary
}

main "$@"
