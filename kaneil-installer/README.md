# KaNeil (Pelican) Installer

One-command installer for the Pelican game server management platform — **Panel + Wings, both in Go**.

## Quick Start

```bash
bash <(curl -s https://raw.githubusercontent.com/YanIanZ/pelican-go/main/kaneil-installer/install.sh)
```

This single command detects your OS, installs all dependencies, builds both Panel and Wings from Go source, and configures everything.

## Requirements

| Component | Minimum |
|-----------|---------|
| OS | Ubuntu 24+ / Debian 11+ / Rocky 9 / AlmaLinux 9 |
| RAM | 2 GB |
| Disk | 10 GB |

Everything else (Go, MariaDB, Redis, Docker, Nginx) is installed automatically.

## What It Installs

| Layer | Tool | Location |
|-------|------|----------|
| Language | Go 1.25 | `/usr/local/go` |
| Database | MariaDB 10.11+ | system |
| Cache | Redis 7 | system |
| Proxy | Nginx | system |
| Container | Docker | system |
| **Panel** | Go binary | `/opt/pelican/panel` |
| **Panel config** | YAML | `/etc/pelican/panel.yml` |
| **Wings** | Go binary | `/opt/pelican/wings` |
| **Wings config** | YAML | `/etc/pelican/config.yml` |

## Options

All configurable via environment variables:

```bash
# Non-interactive full install with custom settings:
FQDN=game.example.com                                \
MYSQL_DB=pelican                                     \
MYSQL_USER=pelican                                   \
MYSQL_PASSWORD=mysecret                              \
CONFIGURE_FIREWALL=yes                               \
sudo -E bash <(curl -s https://raw.githubusercontent.com/YanIanZ/pelican-go/main/kaneil-installer/install.sh)
```

| Variable | Default | Description |
|----------|---------|-------------|
| `FQDN` | `$(hostname -f)` | Panel domain/IP |
| `MYSQL_DB` | `pelican` | Database name |
| `MYSQL_USER` | `pelican` | Database user |
| `MYSQL_PASSWORD` | auto-generated | Database password |
| `APP_SECRET` | auto-generated | JWT signing key |
| `INSTALL_PANEL` | `yes` | Install the Panel |
| `INSTALL_WINGS` | `yes` | Install Wings daemon |
| `INSTALL_NGINX` | `yes` | Install Nginx reverse proxy |
| `CONFIGURE_FIREWALL` | `no` | Open ports 80/443/2022/8080 |
| `ASSUME_SSL` | `no` | Configure Let's Encrypt SSL |
| `GO_VERSION` | `1.25.7` | Go version to install |

### Wings-only install

```bash
INSTALL_PANEL=no sudo -E bash <(curl -s https://raw.githubusercontent.com/YanIanZ/pelican-go/main/kaneil-installer/install.sh)
```

### Panel-only install

```bash
INSTALL_WINGS=no sudo -E bash <(curl -s https://raw.githubusercontent.com/YanIanZ/pelican-go/main/kaneil-installer/install.sh)
```

## After Install

```
Panel:  http://<FQDN>
Wings:  systemctl start pelican-wings  (after setting token)
```

1. Open the Panel URL in your browser
2. Create an admin user
3. Add a **Node** in the admin panel → copy `uuid`, `token_id`, `token`
4. Paste into `/etc/pelican/config.yml`
5. Start Wings: `systemctl start pelican-wings`

## Service Management

```bash
# Panel
systemctl status pelican-panel
journalctl -u pelican-panel -f

# Wings
systemctl status pelican-wings
journalctl -u pelican-wings -f

# Restart
systemctl restart pelican-panel
systemctl restart pelican-wings
```

## Ports

| Port | Service |
|------|---------|
| 80/443 | Nginx → Panel |
| 8080 | Panel (internal) / Wings API |
| 2022 | Wings SFTP |
| 3306 | MariaDB (localhost only) |
| 6379 | Redis (localhost only) |

## Uninstall

```bash
systemctl stop pelican-panel pelican-wings
systemctl disable pelican-panel pelican-wings
rm -f /etc/systemd/system/pelican-panel.service /etc/systemd/system/pelican-wings.service
rm -rf /opt/pelican /etc/pelican
systemctl daemon-reload
```

## License

GPL-3.0 — see [LICENSE](../LICENSE)
