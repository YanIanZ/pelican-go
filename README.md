# Pelican (Go)

Game server management platform — **Panel + Wings, fully in Go**. No PHP.

## One-Command Install

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/YanIanZ/pelican-go/main/install.sh)
```

Detects OS, installs deps, builds Panel + Wings from source, starts services.

## What's Inside

```
pelican-go/
├── install.sh      # One-command installer
├── panel/          # Go admin panel (Gin, GORM, HTMX)
├── wings/          # Go daemon agent (Docker, SFTP, WebSocket)
└── kaneil-installer/  # Reference installer (legacy)
```

| Component | Language | Stack | DB |
|-----------|----------|-------|-----|
| Panel | Go | Gin + GORM + HTMX + Alpine.js | MySQL/Postgres |
| Wings | Go | Gin + Cobra + GORM + Docker | SQLite |

## Manual Install

```bash
git clone https://github.com/YanIanZ/pelican-go.git
cd pelican-go
sudo ./install.sh
```

## Options

```bash
FQDN=game.example.com MYSQL_PASSWORD=secret sudo -E ./install.sh
```

| Variable | Default | Description |
|----------|---------|-------------|
| `FQDN` | `$(hostname -f)` | Panel domain |
| `MYSQL_PASSWORD` | auto-generated | DB password |
| `INSTALL_PANEL` | `yes` | Install Panel |
| `INSTALL_WINGS` | `yes` | Install Wings |
| `CONFIGURE_FIREWALL` | `no` | Open ports |
| `ASSUME_SSL` | `no` | Let's Encrypt |
| `GO_VERSION` | `1.25.7` | Go version |

## After Install

1. Open `http://<FQDN>`
2. Create admin user
3. Add a Node → copy `uuid`, `token_id`, `token`
4. Paste into `/etc/pelican/config.yml`
5. `systemctl start pelican-wings`

```bash
# Panel
systemctl status pelican-panel
journalctl -u pelican-panel -f

# Wings
systemctl start pelican-wings
journalctl -u pelican-wings -f
```

## License

MIT (Wings) / AGPL-3.0 (Panel)
