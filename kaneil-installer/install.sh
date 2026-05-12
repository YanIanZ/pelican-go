#!/bin/bash
set -e

######################################################################################
#  KaNeil (Pelican) — Full Go Installer                                              #
#  One-command Panel + Wings installation from source                                #
#  github.com/YanIanZ/pelican-go                                                     #
######################################################################################

if [[ $EUID -ne 0 ]]; then
    echo "ERROR: Must run as root." >&2
    exit 1
fi

# ---- Configurable env vars ----
FQDN="${FQDN:-localhost}"
EMAIL="${EMAIL:-}"
ASSUME_SSL="${ASSUME_SSL:-false}"
CONFIGURE_LETSENCRYPT="${CONFIGURE_LETSENCRYPT:-false}"
CONFIGURE_FIREWALL="${CONFIGURE_FIREWALL:-false}"
MYSQL_DB="${MYSQL_DB:-panel}"
MYSQL_USER="${MYSQL_USER:-kaneil}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-$(head -c 32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9')}"
PANEL_TAG="${PANEL_TAG:-main}"
WINGS_TAG="${WINGS_TAG:-main}"
INSTALL_PANEL="${INSTALL_PANEL:-true}"
INSTALL_WINGS="${INSTALL_WINGS:-true}"
GO_VERSION="${GO_VERSION:-1.25.7}"
INSTALL_DIR="${INSTALL_DIR:-/opt/pelican}"
REPO_URL="${REPO_URL:-https://github.com/YanIanZ/pelican-go.git}"
BRANCH="${BRANCH:-main}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$SCRIPT_DIR/.."

# Load lib if available
if [ -f "$SCRIPT_DIR/lib/lib.sh" ]; then
    source "$SCRIPT_DIR/lib/lib.sh"
fi

# ---- Colors ----
C_NC='\033[0m'; C_G='\033[0;32m'; C_Y='\033[1;33m'; C_R='\033[0;31m'; C_B='\033[0;34m'
OK="  [${C_G}OK${C_NC}]"
WARN="  [${C_Y}!!${C_NC}]"
ERR="[ERROR]"

# ---- Logging ----
log()   { echo -e "${C_B}  >>>${C_NC} $*"; }
good()  { echo -e "$OK $*"; }
warn()  { echo -e "$WARN $*"; }
fail()  { echo -e "${C_R}${ERR}${C_NC} $*" >&2; }
banner() {
    echo ""
    echo -e "${C_B}  ╔══════════════════════════════════════════╗${C_NC}"
    echo -e "${C_B}  ║      ${C_G}KaNeil (Pelican) Go Installer${C_B}       ║${C_NC}"
    echo -e "${C_B}  ║    Panel + Wings — Full Go Source Build   ║${C_NC}"
    echo -e "${C_B}  ╚══════════════════════════════════════════╝${C_NC}"
    echo ""
}

# ---- OS Detection ----
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS_ID="$(echo "$ID" | tr '[:upper:]' '[:lower:]')"
        OS_VER="$(echo "$VERSION_ID" | cut -d. -f1)"
    elif [ -f /etc/debian_version ]; then
        OS_ID="debian"
        OS_VER="$(cat /etc/debian_version | cut -d. -f1)"
    else
        fail "Unsupported OS"
        exit 1
    fi
    case "$(uname -m)" in
        x86_64)  ARCH="amd64";;
        aarch64|arm64) ARCH="arm64";;
        *) fail "Unsupported arch: $(uname -m)"; exit 1;;
    esac
    log "OS: $OS_ID $OS_VER | Arch: $ARCH"
}

# ---- Package helpers ----
pkg_install() {
    case "$OS_ID" in
        ubuntu|debian)
            export DEBIAN_FRONTEND=noninteractive
            apt-get update -qq && apt-get install -y -qq "$@"
            ;;
        rocky|almalinux|rhel|centos|fedora)
            dnf install -y -q "$@"
            ;;
    esac
}

# ---- Install Go ----
install_go() {
    if command -v go &>/dev/null; then
        local v; v=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | head -1 | tr -d 'go')
        if printf '%s\n%s\n' "1.25" "$v" | sort -V -C 2>/dev/null; then
            good "Go $v found"
            return 0
        fi
    fi

    log "Installing Go $GO_VERSION..."
    local go_tar="go${GO_VERSION}.linux-${ARCH}.tar.gz"
    curl -sSfL "https://go.dev/dl/${go_tar}" -o "/tmp/${go_tar}"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "/tmp/${go_tar}"
    rm -f "/tmp/${go_tar}"

    grep -q '/usr/local/go/bin' /etc/profile 2>/dev/null || \
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    export PATH="$PATH:/usr/local/go/bin"
    good "Go $GO_VERSION installed"
}

# ---- Install Docker ----
install_docker() {
    if command -v docker &>/dev/null; then
        good "Docker found"
        return 0
    fi
    log "Installing Docker..."
    curl -fsSL https://get.docker.com | bash
    systemctl enable --now docker
    good "Docker installed"
}

# ---- Install MariaDB ----
install_mariadb() {
    if command -v mariadb &>/dev/null; then
        good "MariaDB found"
        return 0
    fi
    if command -v mysql &>/dev/null; then
        good "MySQL found"
        return 0
    fi
    log "Installing MariaDB..."
    case "$OS_ID" in
        ubuntu|debian)
            pkg_install mariadb-server mariadb-client
            ;;
        rocky|almalinux|rhel|centos|fedora)
            curl -sS https://downloads.mariadb.com/MariaDB/mariadb_repo_setup | bash
            pkg_install MariaDB-server MariaDB-client
            ;;
    esac
    systemctl enable --now mariadb || systemctl enable --now mysql
    good "MariaDB installed"
}

# ---- Install Redis ----
install_redis() {
    if command -v redis-server &>/dev/null; then
        good "Redis found"
        return 0
    fi
    log "Installing Redis..."
    pkg_install redis-server 2>/dev/null || pkg_install redis
    systemctl enable --now redis-server 2>/dev/null || systemctl enable --now redis
    good "Redis installed"
}

# ---- Install Nginx ----
install_nginx() {
    if command -v nginx &>/dev/null; then
        good "Nginx found"
        return 0
    fi
    log "Installing Nginx..."
    pkg_install nginx
    systemctl enable --now nginx
    good "Nginx installed"
}

# ---- Clone/Build Panel from Go source ----
build_panel() {
    log "Building Panel from Go source..."
    local panel_src="$REPO_DIR/panel"

    if [ ! -f "$panel_src/go.mod" ]; then
        log "Cloning panel source..."
        mkdir -p "$(dirname "$panel_src")"
        if [ -d "$panel_src" ]; then rm -rf "$panel_src"; fi
        git clone --branch "$PANEL_TAG" "$REPO_URL" /tmp/pelican-clone
        cp -r /tmp/pelican-clone/panel "$panel_src"
        rm -rf /tmp/pelican-clone
    fi

    cd "$panel_src"
    go mod download
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/panel" ./cmd/panel
    good "Panel binary: $INSTALL_DIR/panel"
}

# ---- Build Wings from Go source ----
build_wings() {
    log "Building Wings from Go source..."
    local wings_src="$REPO_DIR/wings"

    if [ ! -f "$wings_src/go.mod" ]; then
        log "Cloning wings source..."
        git clone --branch "$WINGS_TAG" "$REPO_URL" /tmp/pelican-clone
        cp -r /tmp/pelican-clone/wings "$wings_src"
        rm -rf /tmp/pelican-clone
    fi

    cd "$wings_src"
    go mod download
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/wings" .
    good "Wings binary: $INSTALL_DIR/wings"
}

# ---- Setup Panel database ----
setup_database() {
    log "Setting up database..."
    mariadb -u root -e "CREATE DATABASE IF NOT EXISTS $MYSQL_DB;" 2>/dev/null || true
    mariadb -u root -e "CREATE USER IF NOT EXISTS '$MYSQL_USER'@'127.0.0.1' IDENTIFIED BY '$MYSQL_PASSWORD';" 2>/dev/null || true
    mariadb -u root -e "ALTER USER '$MYSQL_USER'@'127.0.0.1' IDENTIFIED BY '$MYSQL_PASSWORD';" 2>/dev/null || true
    mariadb -u root -e "GRANT ALL PRIVILEGES ON $MYSQL_DB.* TO '$MYSQL_USER'@'127.0.0.1' WITH GRANT OPTION;" 2>/dev/null || true
    mariadb -u root -e "FLUSH PRIVILEGES;" 2>/dev/null || true
    good "Database $MYSQL_DB ready (user: $MYSQL_USER)"
}

# ---- Write Panel config ----
write_panel_config() {
    log "Writing panel config..."
    mkdir -p /etc/pelican
    cat > /etc/pelican/panel.yml << YEOF
app:
  name: Pelican
  env: production
  debug: false
  url: "${PROTO:-http}://${FQDN}"
  timezone: UTC
  locale: en

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
YEOF
    good "Panel config: /etc/pelican/panel.yml"
    log "MySQL password: ${MYSQL_PASSWORD}"
}

# ---- Write Wings config ----
write_wings_config() {
    log "Writing wings config..."
    mkdir -p /etc/pelican
    mkdir -p /var/lib/pelican/volumes /var/log/pelican /tmp/pelican

    local fqdn_lower
    fqdn_lower=$(echo "$FQDN" | tr '[:upper:]' '[:lower:]')

    cat > /etc/pelican/wings.yml << YEOF
debug: false
app_name: Pelican
uuid: ""
token_id: ""
token: ""
api:
  host: 0.0.0.0
  port: 8080
  ssl:
    enabled: $([ "$ASSUME_SSL" = true ] && echo "true" || echo "false")
    cert: /etc/letsencrypt/live/${fqdn_lower}/fullchain.pem
    key: /etc/letsencrypt/live/${fqdn_lower}/privkey.pem
  upload_limit: 100
system:
  data: /var/lib/pelican/volumes
  sftp:
    bind_port: 2022
    bind_address: 0.0.0.0
  log_directory: /var/log/pelican
  tmp_directory: /tmp/pelican
allowed_mounts: []
remote: "${PROTO:-http}://${FQDN}"
docker:
  network:
    name: pelican_nw
    interface: 172.18.0.1
    dns: ["1.1.1.1", "8.8.8.8"]
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
  max_crashes: 3
  timeout: 60
backups:
  write_limit: 0
  compression_level: best_speed
transfers:
  download_limit: 0
YEOF
    good "Wings config: /etc/pelican/wings.yml"
}

# ---- Nginx config ----
write_nginx_config() {
    if [ ! -f "/etc/nginx/sites-available/kaneil.conf" ]; then
        log "Writing Nginx config..."
        mkdir -p /etc/nginx/sites-available /etc/nginx/sites-enabled

        if [ "$ASSUME_SSL" = true ]; then
            cat > /etc/nginx/sites-available/kaneil.conf << 'NGXEOF'
server {
    listen 80;
    server_name CHANGE_ME;
    return 301 https://$host$request_uri;
}
server {
    listen 443 ssl http2;
    server_name CHANGE_ME;
    root /var/www/kaneil/public;
    index index.html index.htm;

    ssl_certificate /etc/letsencrypt/live/CHANGE_ME/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/CHANGE_ME/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api/client/servers {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
NGXEOF
        else
            cat > /etc/nginx/sites-available/kaneil.conf << 'NGXEOF'
server {
    listen 80;
    server_name CHANGE_ME;
    root /var/www/kaneil/public;
    index index.html index.htm;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api/client/servers {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
NGXEOF
        fi
        sed -i "s/CHANGE_ME/${FQDN}/g" /etc/nginx/sites-available/kaneil.conf
        ln -sf /etc/nginx/sites-available/kaneil.conf /etc/nginx/sites-enabled/kaneil.conf
        rm -f /etc/nginx/sites-enabled/default
        systemctl restart nginx
        good "Nginx configured for $FQDN"
    else
        good "Nginx config exists, skipping"
    fi
}

# ---- Setup systemd ----
setup_panel_service() {
    log "Installing panel systemd service..."
    cat > /etc/systemd/system/pelican-panel.service << SRVEOF
[Unit]
Description=Pelican Panel
After=network.target mariadb.service mysql.service redis.service redis-server.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/panel --config=/etc/pelican/panel.yml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
SRVEOF
    systemctl daemon-reload
    systemctl enable pelican-panel
    systemctl start pelican-panel 2>/dev/null || warn "Panel service failed to start (check /etc/pelican/panel.yml)"
    good "Panel service installed"
}

setup_wings_service() {
    log "Installing wings systemd service..."
    cat > /etc/systemd/system/pelican-wings.service << SRVEOF
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
SRVEOF
    systemctl daemon-reload
    systemctl enable pelican-wings
    good "Wings service installed"
    warn "Start Wings after setting uuid/token_id/token in /etc/pelican/wings.yml"
}

# ---- Firewall ----
setup_firewall() {
    log "Configuring firewall..."
    if command -v ufw &>/dev/null; then
        ufw allow 80/tcp  2>/dev/null || true
        ufw allow 443/tcp 2>/dev/null || true
        ufw allow 2022/tcp 2>/dev/null || true
        ufw allow 8080/tcp 2>/dev/null || true
        ufw --force reload
    elif command -v firewall-cmd &>/dev/null; then
        firewall-cmd --zone=public --add-port=80/tcp --permanent
        firewall-cmd --zone=public --add-port=443/tcp --permanent
        firewall-cmd --zone=public --add-port=2022/tcp --permanent
        firewall-cmd --zone=public --add-port=8080/tcp --permanent
        firewall-cmd --reload
    fi
    good "Firewall configured"
}

# ---- Main ----
main() {
    banner

    detect_os

    log "Installing dependencies..."
    install_go
    pkg_install curl git tar certbot

    if [ "$INSTALL_PANEL" = true ]; then
        log ""
        log "====== Panel Installation ======"
        mkdir -p "$INSTALL_DIR"
        install_mariadb
        install_redis
        setup_database
        build_panel
        write_panel_config
        setup_panel_service
    fi

    if [ "$INSTALL_WINGS" = true ]; then
        log ""
        log "====== Wings Installation ======"
        mkdir -p "$INSTALL_DIR"
        install_docker
        build_wings
        write_wings_config
        setup_wings_service
    fi

    install_nginx
    write_nginx_config

    if [ "$CONFIGURE_FIREWALL" = true ]; then
        setup_firewall
    fi

    echo ""
    echo -e "${C_G}════════════════════════════════════════${C_NC}"
    echo -e "${C_G}     Installation Complete!${C_NC}"
    echo -e "${C_G}════════════════════════════════════════${C_NC}"
    echo ""
    if [ "$INSTALL_PANEL" = true ]; then
        echo "  Panel:    http://${FQDN}"
        echo "  Service:  systemctl status pelican-panel"
        echo "  Config:   /etc/pelican/panel.yml"
        echo "  Logs:     journalctl -u pelican-panel -f"
        echo ""
        echo "  DB name:     ${MYSQL_DB}"
        echo "  DB user:     ${MYSQL_USER}"
        echo "  DB password: ${MYSQL_PASSWORD}"
        echo ""
    fi
    if [ "$INSTALL_WINGS" = true ]; then
        echo "  Wings:    systemctl status pelican-wings"
        echo "  Config:   /etc/pelican/wings.yml"
        echo "  Logs:     journalctl -u pelican-wings -f"
        echo ""
        echo "  IMPORTANT: Edit /etc/pelican/wings.yml with node uuid,"
        echo "  token_id, and token from Panel before starting Wings:"
        echo "    systemctl start pelican-wings"
        echo ""
    fi
}

main "$@"
