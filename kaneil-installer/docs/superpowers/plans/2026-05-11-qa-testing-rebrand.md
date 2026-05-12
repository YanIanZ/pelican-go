# QA Testing Plan: KaNeil Rebrand (egg→map, wings→ship)
**Date:** 2026-05-11  
**Executor:** QA/Deployment team  
**Duration:** 4–6 hours  
**Blocker:** All fixes must be deployed to respective branches

---

## Scope
Verify code review fixes work in realistic deployment scenarios across 4 OS targets.

---

## Test Matrix

### Fresh Install Tests

#### Test 1: Ubuntu 24.04 + Panel + Ship (www-data)
**Setup:** 
- Launch EC2 t3.medium Ubuntu 24.04, attach 50GB root
- Assign elastic IP, configure DNS to `panel.test.example.com`
- SSH as ubuntu user

**Commands:**
```bash
sudo su -
export FQDN=panel.test.example.com
export email=test@example.com
export user_email=admin@example.com
export user_username=admin
export user_password=$(openssl rand -base64 32)
export CONFIGURE_LETSENCRYPT=false
export ASSUME_SSL=true

curl -sSL https://raw.githubusercontent.com/YanIanZ/KaNeil-Installer/main/install.sh | bash
```

**Validations:**
- [ ] Installer completes without error
- [ ] www-data owns /var/www/kaneil
- [ ] `systemctl status ship` active
- [ ] `systemctl status kaneil` (queue worker) active
- [ ] `curl -k https://localhost/` returns 200 (or 302 to login)
- [ ] Admin user login works (`$user_username` / `$user_password`)

**Expected:** Panel loads, queue worker running, storage/ writable by www-data

---

#### Test 2: Debian 12 + Panel + Ship (www-data)
**Repeat Test 1 with Debian 12 AMI**

**Expected:** Identical behavior to Ubuntu

---

#### Test 3: Rocky 9 + Panel + Ship (nginx)
**Setup:**
- Launch EC2 t3.medium Rocky 9, attach 50GB root
- Same FQDN, email env vars

**Commands:**
- Same installer URL

**Validations:**
- [ ] nginx owns /var/www/kaneil
- [ ] PHP-FPM detected and restarted (check `systemctl status php-fpm`)
- [ ] Smoke test passed in installer log (HTTP 200)
- [ ] Same login/queue/storage checks as Test 1

**Expected:** nginx user ownership, no www-data errors

---

#### Test 4: AlmaLinux 9 + Panel + Ship (nginx)
**Repeat Test 3 with AlmaLinux 9 AMI**

**Critical Validation (AlmaLinux case fix):**
- [ ] `grep OS= /etc/os-release` returns `OS=almalinux`
- [ ] lib.sh detected OS correctly (check installer log: "Detected OS: almalinux")
- [ ] All deps installed (php, nginx, mariadb, redis)
- [ ] `/etc/nginx/conf.d/kaneil.conf` exists and symlinked correctly
- [ ] Installer didn't fall through to unknown OS path

**Expected:** No "Could not determine OS" errors, full install

---

### Map Import Tests

#### Test 5: Bulk Map Import (parkervcp/eggs)
**Setup:** Use Test 1 (Ubuntu panel) or Test 2 (Debian panel)

**Commands:**
```bash
cd /var/www/kaneil
git clone --depth 1 https://github.com/parkervcp/eggs storage/eggs/parkervcp-eggs
sudo -u www-data php artisan p:map:import-bulk storage/eggs/parkervcp-eggs

# Verify
php artisan tinker
App\Models\Map::count()
```

**Validations:**
- [ ] Import completes without "email validation" errors
- [ ] Maps created with startup_commands, docker_images normalized
- [ ] `Map::count()` > 0
- [ ] Sample map has `MapVariable` records (check variables relationship)

**Expected:** 100+ maps imported, non-email authors normalized to 'unknown@kaneil.dev'

---

#### Test 6: Server Creation from Imported Map
**Setup:** Continue from Test 5

**Commands:**
```bash
# Via panel web UI:
# 1. Admin > Maps > select "Minecraft Forge" or similar
# 2. Servers > Create Server
# 3. Select map, fill hostname/instance type
# 4. Submit

# OR via artisan:
php artisan tinker
$map = App\Models\Map::first();
$env = $map->variables->mapWithKeys(fn($v) => [$v->env_variable => $v->default_value])->toArray();
App\Models\Server::create([
  'map_id' => $map->id,
  'name' => 'test-forge',
  'description' => 'Test Forge Server',
  'environment' => $env,
  'memory_limit' => 2048,
  'disk_limit' => 10240,
  'cpu_limit' => 200,
  'threads' => 4,
  'oom_killer' => true,
]);
```

**Validations:**
- [ ] Server created without validation errors
- [ ] Environment array backfilled from MapVariable defaults
- [ ] Server status page loads
- [ ] Can view server details, edit configuration

**Expected:** Servers created with backfilled env, no validation rejection

---

### Update Tests

#### Test 7: Update Panel (composer failure rollback)
**Setup:** Use Test 1 (Ubuntu panel)

**Commands:**
```bash
# Corrupt composer.json to trigger failure
cd /var/www/kaneil
sed -i 's/"laravel\/framework".*/"laravel\/framework": "99.99.99"/' composer.json

# Run update (should rollback)
bash /tmp/update-panel.sh

# Verify rollback
systemctl status kaneil
curl -k https://localhost/
```

**Validations:**
- [ ] Update detected composer error
- [ ] Backup created at `/var/www/kaneil_backup_<timestamp>`
- [ ] Rollback restores from backup (rsync used)
- [ ] .env preserved (check APP_URL, DB_PASSWORD not lost)
- [ ] Panel still accessible after rollback (no broken artisan)

**Expected:** Rollback succeeds, panel recovers with intact .env and storage/

---

#### Test 8: Update Ship (binary atomic)
**Setup:** Use Test 1 (Ubuntu ship)

**Commands:**
```bash
# Check current version
/usr/local/bin/ship --version

# Simulate interrupted download (kill curl mid-transfer)
# Run update in background, kill during download
bash /tmp/update-ship.sh &
sleep 3
pkill -f "curl.*ship"

# Check binary
ls -lh /usr/local/bin/ship
/usr/local/bin/ship --version || echo "Binary corrupted"

# Retry update
bash /tmp/update-ship.sh
```

**Validations:**
- [ ] After interrupt, /usr/local/bin/ship still executable (not corrupted)
- [ ] /tmp/ship_latest cleaned up (not left hanging)
- [ ] Retry succeeds and updates binary
- [ ] ship --version shows new version

**Expected:** Atomic download prevents corruption, partial files never reach systemd path

---

#### Test 9: Update Ship (version output)
**Setup:** Continue from Test 8

**Commands:**
```bash
bash /tmp/update-ship.sh | grep "Ship version"
```

**Validations:**
- [ ] Version output shows actual version (e.g., "v2.0.0")
- [ ] No garbage output or "undefined" strings

**Expected:** Clean version output, no stale probe artifacts

---

### Smoke Test Validation

#### Test 10: Smoke Test Regex & Host Header
**Setup:** Use Test 1 or Test 2 (panel)

**Commands:**
```bash
# Manually run smoke test with wrong Host header
curl -sk -o /dev/null -w "%{http_code}" -H "Host: wrong.host.local" https://127.0.0.1/
# Should return 404 or 502 (not 200)

# Run with correct Host header
curl -sk -o /dev/null -w "%{http_code}" -H "Host: panel.test.example.com" https://127.0.0.1/
# Should return 200 or 302 (login redirect)
```

**Validations:**
- [ ] Wrong host returns non-2xx/3xx (4xx/5xx)
- [ ] Correct host returns 2xx/3xx
- [ ] Installer smoke test would correctly fail on wrong host

**Expected:** Host header validation prevents false positives

---

### OS Detection Tests

#### Test 11: Detect Web User (Ubuntu vs Rocky)
**Setup:** Use Test 1 & Test 3 (Ubuntu & Rocky)

**Commands (Ubuntu):**
```bash
# Check fix-all.sh web user detection
bash fix-all.sh 2>&1 | grep "WEB_USER="
# Should detect www-data
```

**Commands (Rocky):**
```bash
bash fix-all.sh 2>&1 | grep "WEB_USER="
# Should detect nginx
```

**Validations:**
- [ ] Ubuntu: WEB_USER=www-data
- [ ] Rocky: WEB_USER=nginx
- [ ] Permissions applied to correct user

**Expected:** Dynamic detection works, no hardcoded user mismatches

---

#### Test 12: PHP-FPM Restart Loop (8.5/8.4/8.3/8.2)
**Setup:** Use Test 1 (Ubuntu) or Test 3 (Rocky)

**Commands:**
```bash
bash fix-all.sh 2>&1 | grep -A 2 "PHP-FPM"
# Should see "Restarted php8.5-fpm" or similar
```

**Validations:**
- [ ] Correct PHP version detected and restarted
- [ ] OPcache cleared (systemctl restart php-fpm)
- [ ] No errors about missing version

**Expected:** Correct PHP-FPM service restarted, OPcache cleared

---

### Dashboard/Panel Tests

#### Test 13: Galleon UI Load & Map Management
**Setup:** Use Test 1 (Ubuntu panel)

**Commands (via browser):**
1. Navigate to https://panel.test.example.com
2. Login as admin
3. Go to Admin > Maps
4. List maps (should show imported ones)
5. Click map details (view variables, docker_images)
6. Go to Admin > Servers
7. Create new server (select map, fill required fields)
8. View created server

**Validations:**
- [ ] Galleon UI loads without JS errors (browser console clean)
- [ ] Maps list shows imported maps with correct names
- [ ] Variables display (check MapVariable.label, description, env_variable)
- [ ] Server creation form auto-populates from map template
- [ ] Server created with correct environment variables

**Expected:** Galleon UI functional, map/server management works, no rebrand terminology issues visible

---

## Rollback Plan

If any test fails:
1. **Panel rollback:** Use backup at `/var/www/kaneil_backup_<timestamp>` (created by update-panel.sh)
2. **Ship rollback:** `/usr/local/bin/ship.backup` exists (created by update-ship.sh)
3. **Full rollback:** Terminate instance, revert branch, re-test

---

## Pass Criteria

All 13 tests must pass with zero critical failures:
- ✅ All 4 OS installs complete without error
- ✅ Map import succeeds, non-email authors handled
- ✅ Server creation backfills env from MapVariable defaults
- ✅ Update rollback preserves .env, storage/
- ✅ Binary downloads are atomic
- ✅ Smoke test validates Host header + HTTP status
- ✅ Web user/PHP-FPM detected dynamically
- ✅ Galleon UI functional

---

## Sign-off

- [ ] QA Lead: ____________ Date: ___________
- [ ] Deployment Lead: ____________ Date: ___________
- [ ] Ready for production: YES / NO

