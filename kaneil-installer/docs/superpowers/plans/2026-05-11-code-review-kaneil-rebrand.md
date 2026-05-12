# KaNeil Rebrand Code Review & Fix Plan
**Date:** 2026-05-11  
**Status:** Complete  
**Scope:** Panel (experimental/v2.0-EX), Ship (Go), Installer (bash)

---

## Objectives
- Identify CRIT/WARN/NIT issues in rebrand implementation (egg→map, wings→ship terminology)
- Validate map import/export pipeline, variable backfill, docker_images normalization
- Verify installer atomic operations, backup/rollback safety, OS detection
- Fix all CRIT + WARN issues before production deployment

---

## Issues Found & Fixed

### Panel (3 CRIT)

| Issue | Impact | Fix |
|-------|--------|-----|
| Maps import validation incomplete (ship_id, uuid, startup_commands missing) | Bulk imports fail silently | Built full Map payload, FirstOrCreate Default ship, MapVariable creation |
| Server env variables not backfilled from MapVariable defaults | Server creation rejects valid maps | Extended validation, backfill env before ServerCreationService |
| Non-email author strings fail email validation | Egg imports silently fail | Added fallback normalization to 'unknown@kaneil.dev' |

**Commits:** 063f0b940 (galleon validation), a06f0b940 (rebrand completion)

### Installer (1 CRIT + 8 WARN)

**CRIT:** AlmaLinux case mismatch (OS=almalinux but case matched only rocky|alma) → no deps, no nginx, no perms  
**Fix:** Changed all case branches to `rocky | almalinux | alma`

**WARN Issues Fixed:**
1. Backup/rollback corruption (.env not restored) → rsync --delete, verify backup integrity
2. Smoke test false positives (4xx accepted, no Host header) → regex [23][0-9][0-9], extract HOST from APP_URL
3. Binary download non-atomic (partial file possible) → stage /tmp/ship.new, size check, atomic mv
4. Stale version probe (undefined var executed) → removed broken probe
5. Hardcoded nginx user breaks Ubuntu → dynamic detection (id -u www-data)
6. PHP-FPM hardcoded to 8.5 only → loop 8.5/8.4/8.3/8.2/php-fpm
7. Release URL pinned to stable → changed to experimental-latest, PANEL_DL_URL override
8. APP_URL not displayed → grep .env, show in final output

**Commits:** b9477fa (post-review fixes)

### Ship (Go)
Clean review, no issues found.

---

## Testing Checklist

- [ ] Ubuntu 24.04 fresh install (panel + ship)
- [ ] Debian 12 fresh install (panel + ship)
- [ ] Rocky 9 fresh install (panel + ship)
- [ ] AlmaLinux 9 fresh install (panel + ship)
- [ ] Map import from parkervcp/eggs (game + application maps)
- [ ] Server create with required variables (backfill env from defaults)
- [ ] Update-panel.sh on broken install (backup/rollback verify)
- [ ] Update-ship.sh version output
- [ ] fix-all.sh end-to-end (web user, PHP-FPM, APP_URL output)
- [ ] Smoke test HTTP check with wrong Host: header (should fail)

---

## Deployment Readiness

**Pre-flight:**
- [x] All CRIT issues fixed
- [x] All WARN issues fixed
- [x] Ship binary clean
- [x] Code review complete
- [ ] QA testing on target OSes

**Release:**
- Panel: `experimental/v2.0-EX` branch → `experimental-latest` GitHub release (auto via CI)
- Installer: `main` branch → merged, used in `experimental-latest` release tag
- Ship: `main` branch → Go binary in `experimental-latest` release

---

## Follow-up Actions

1. **QA Testing** — Run checklist above on fresh instances across all supported OSes
2. **NIT Resolution** — Address branding text, comments in future v2.0 stable PR
3. **Stable v2.0** — Update release URLs from experimental-latest to releases/latest tag
4. **Monitoring** — Watch production installs for edge cases (Docker image normalization, MapVariable backfill failures)

---

## Known Limitations

- Experimental v2.0-EX uses `experimental-latest` release tag (unstable channel)
- Map JSON schema validation (docker_images, startup_commands) relies on egg structure consistency
- AlmaLinux detected by case match (handles OS var + user alias, but not exhaustive)
- Backup restore assumes rsync available (fallback to cp -a for systems without rsync)

---

## Architecture Decisions

**Atomic Binary Download:** Stage to /tmp, size-check, mv ensures interrupted download never leaves corrupt binary in systemd path

**Dynamic Web User Detection:** id -u check + fallback covers Ubuntu/Debian (www-data) and Rocky/AlmaLinux (nginx) without hardcoding

**Backup Verification:** Check artisan exists in backup after cp finishes — early detection of corrupted backup before composer failure rollback needed

**Smoke Test Regex:** `^[23][0-9][0-9]$` enforces 2xx/3xx only; optional Host: header prevents nginx default vhost bypass

**MapVariable Backfill:** Load defaults before ServerCreationService invocation — consistent env across server lifecycle, no late binding issues
