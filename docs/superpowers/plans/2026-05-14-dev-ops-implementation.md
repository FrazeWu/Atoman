# DevOps and Environment Consistency Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reproduce, verify, and harden the completed DevOps/environment-consistency migration so local development runs against PostgreSQL + MinIO, runtime configuration no longer falls back to the old SQLite/local-storage path, and the current production packaging path is explicitly validated.

**Architecture:** Treat `plan/dev_ops_task_plan.md` as the completion ledger, `plan/dev_ops_findings.md` as the runtime risk record, and `plan/dev_ops_progress.md` as the evidence of what was already proven in a working local environment. Execute this plan as a verification-first pass: confirm the current repository still matches the documented DevOps design, re-run the critical environment and migration checks, and make any remaining artifact/documentation gaps explicit instead of assuming the old evidence is still valid.

**Tech Stack:** Docker Compose, PostgreSQL 16, MinIO, Go, Gin, GORM, Bun, Vue 3, Nginx, supervisord

---

## Upstream Sources

- `plan/dev_ops_task_plan.md`
- `plan/dev_ops_findings.md`
- `plan/dev_ops_progress.md`
- Supporting runtime files discovered in the worktree:
  - `docker-compose.dev.yml`
  - `docker-compose.prod.yml`
  - `Dockerfile`
  - `manage.sh`
  - `server/cmd/start_server/main.go`
  - `server/migrate_db.go`
  - `server/internal/handlers/songs_handler.go`
  - `server/internal/handlers/albums_handler.go`
  - `server/.env.example`

## Normalized Current State

The upstream DevOps documents agree that the environment-consistency migration is functionally complete: local development should now use Dockerized PostgreSQL and MinIO while the Go backend and Vue frontend continue to run natively on the host for debugging and hot reload. Storage writes are expected to follow `STORAGE_TYPE`, not user role, and the migration path from `dev.sqlite` + local `uploads/` to PostgreSQL + MinIO is implemented in `server/migrate_db.go`.

The most important runtime fix recorded upstream is already visible in `server/cmd/start_server/main.go`: local startup now prefers `.env.dev` before falling back to `.env`, which prevents the old accidental connection to SQLite/local storage when the infrastructure migration has already been performed.

There are also two important mismatches to keep explicit during execution:

1. The task plan says `.env.dev.example` was created, but the tracked template visible in this worktree is `server/.env.example`; no tracked repo-root `.env.dev.example` was found.
2. The findings recommend a future multi-stage production image, but the current `Dockerfile` is still a runtime-only image that expects prebuilt `web/dist` and `server/main` artifacts.

For execution purposes, treat DevOps as **implemented but requiring fresh end-to-end verification and artifact reconciliation**.

## Confirmed Scope To Preserve

- Local development uses **Docker for infrastructure only**: PostgreSQL + MinIO run in containers; Go and Vue run on the host.
- The backend must connect to PostgreSQL in dev, not silently fall back to SQLite.
- Upload/storage behavior must be governed by `STORAGE_TYPE`, not admin-role branching.
- In `STORAGE_TYPE=s3` mode, music/audio/cover uploads must go to MinIO/S3-compatible storage.
- `server/migrate_db.go` remains the migration path for moving legacy SQLite data and local files into PostgreSQL + MinIO.
- MinIO objects used by the frontend must be publicly reachable or otherwise fetchable by the current frontend rendering path.
- Production deployment remains Compose-based today, but its requirements and limitations must be explicit.

## Verified Architecture Decisions From Upstream

- **Hybrid local dev model:** infrastructure in Docker, application processes native on the host.
- **Environment parity priority:** local development should mimic production storage/database behavior as much as possible.
- **Storage routing rule:** `STORAGE_TYPE` is the source of truth for local vs S3/MinIO writes.
- **Migration strategy:** migrate both database rows and local file assets, then rewrite stored `/uploads/...` URLs to the S3/MinIO public prefix.
- **Config risk posture:** avoid ambiguous `.env` fallback in production-like runs; explicitly prefer the dev-specific env source for local development.
- **Production packaging reality:** the current production image path works only after building frontend and backend artifacts ahead of time.

## Known Discrepancies To Resolve During Execution

- `plan/dev_ops_task_plan.md` references `.env.dev.example`, but the visible tracked template is `server/.env.example`.
- `Dockerfile` comments still describe a prebuild workflow, so “production template complete” should be interpreted as “usable current template,” not “fully streamlined immutable build pipeline.”
- `docker-compose.prod.yml` depends on `./server/.env.prod`; that file is environment-specific and may not exist in every checkout, so production verification must separate **image/build verification** from **full deployment verification**.

## Critical Implementation Files To Revalidate

### Planning inputs
- `plan/dev_ops_task_plan.md`
- `plan/dev_ops_findings.md`
- `plan/dev_ops_progress.md`

### Local infrastructure and env configuration
- `docker-compose.dev.yml`
- runtime-only local file: `.env.dev`
- tracked template: `server/.env.example`
- `server/cmd/start_server/main.go`

### Storage and migration path
- `server/internal/handlers/songs_handler.go`
- `server/internal/handlers/albums_handler.go`
- `server/migrate_db.go`

### Production packaging path
- `Dockerfile`
- `docker-compose.prod.yml`
- `manage.sh`

## Task 1: Reconcile the documented DevOps design against the current repository

**Files:**
- Inspect: `plan/dev_ops_task_plan.md`
- Inspect: `plan/dev_ops_findings.md`
- Inspect: `plan/dev_ops_progress.md`
- Inspect: `docker-compose.dev.yml`
- Inspect: `docker-compose.prod.yml`
- Inspect: `Dockerfile`
- Inspect: `manage.sh`
- Inspect: `server/cmd/start_server/main.go`
- Inspect: `server/migrate_db.go`
- Inspect: `server/internal/handlers/songs_handler.go`
- Inspect: `server/internal/handlers/albums_handler.go`
- Inspect: `server/.env.example`

- [ ] **Step 1: Review the upstream completion claims**

Read the three upstream plan files and record these exact buckets in working notes:

```text
- what is claimed complete in phases 1-5
- which runtime risks are still called out in findings
- what was proven by the previous manual verification session
```

Expected: a short reconciliation list that distinguishes “documented complete” from “needs fresh revalidation now.”

- [ ] **Step 2: Confirm the repo still reflects the documented local-dev model**

Inspect these files and verify the expected responsibilities:

```text
- docker-compose.dev.yml → postgres, db-init, minio, minio-setup only
- server/cmd/start_server/main.go → prefers .env.dev before .env
- server/internal/handlers/songs_handler.go → STORAGE_TYPE=s3 writes through S3 client path
- server/internal/handlers/albums_handler.go → cover uploads follow STORAGE_TYPE, not admin role
- server/migrate_db.go → reads dev.sqlite + uploads, writes PostgreSQL + MinIO, rewrites URLs
```

Expected: the current code still matches the upstream DevOps architecture.

- [ ] **Step 3: Record source-of-truth mismatches before running anything**

Write down the discrepancies you can already prove from the repository:

```text
- tracked env template found: server/.env.example
- tracked repo-root .env.dev.example not found in this worktree
- current Dockerfile still requires prebuilt assets/binary
```

Expected: later verification does not silently assume the wrong config/template artifact exists.

## Task 2: Verify local infrastructure bootstrap with PostgreSQL and MinIO

**Files:**
- Use: `docker-compose.dev.yml`
- Use: `.env.dev`
- Reference: `server/.env.example`

- [ ] **Step 1: Confirm the local dev env file exists and contains the required keys**

Run from the repository root:

```bash
python - <<'PY'
from pathlib import Path
required = [
    'POSTGRES_USER', 'POSTGRES_PASSWORD', 'POSTGRES_DB',
    'MINIO_ROOT_USER', 'MINIO_ROOT_PASSWORD',
    'DATABASE_TYPE', 'DATABASE_URL',
    'STORAGE_TYPE', 'S3_BUCKET', 'S3_ENDPOINT', 'S3_URL_PREFIX',
    'AWS_ACCESS_KEY_ID', 'AWS_SECRET_ACCESS_KEY',
    'JWT_SECRET'
]
path = Path('.env.dev')
if not path.exists():
    raise SystemExit('MISSING:.env.dev')
values = {}
for line in path.read_text().splitlines():
    line = line.strip()
    if not line or line.startswith('#') or '=' not in line:
        continue
    k, v = line.split('=', 1)
    values[k] = v
missing = [k for k in required if not values.get(k)]
if missing:
    raise SystemExit('MISSING_KEYS:' + ','.join(missing))
print('OK:.env.dev present with required keys')
print('DATABASE_TYPE=' + values['DATABASE_TYPE'])
print('STORAGE_TYPE=' + values['STORAGE_TYPE'])
PY
```

Expected:

```text
OK:.env.dev present with required keys
DATABASE_TYPE=postgres
STORAGE_TYPE=s3
```

If `DATABASE_TYPE` is `sqlite` or `STORAGE_TYPE` is `local`, stop and fix the env file before continuing.

- [ ] **Step 2: Start the local infrastructure services**

Run:

```bash
docker compose --env-file .env.dev -f docker-compose.dev.yml up -d postgres db-init minio minio-setup
```

Expected: the four infrastructure services are created or reused without compose errors.

- [ ] **Step 3: Confirm service health and init completion**

Run:

```bash
docker compose --env-file .env.dev -f docker-compose.dev.yml ps
```

Expected:

```text
- postgres → healthy
- minio → healthy
- db-init → exited (0)
- minio-setup → exited (0)
```

- [ ] **Step 4: Verify database initialization and bucket setup logs**

Run:

```bash
docker compose --env-file .env.dev -f docker-compose.dev.yml logs db-init minio-setup --tail 100
```

Expected log evidence:

```text
- db-init checked or created the PostgreSQL database named by POSTGRES_DB
- minio-setup created the bucket if missing
- minio-setup ran: mc anonymous set public myminio/$S3_BUCKET
```

- [ ] **Step 5: Verify MinIO health endpoint directly**

Run:

```bash
curl -fsS http://127.0.0.1:9100/minio/health/live
```

Expected: successful response body such as `OK` and exit code 0.

## Task 3: Verify backend and frontend now run against PostgreSQL + MinIO

**Files:**
- Build/Test: `server/cmd/start_server/main.go`
- Build/Test: `server/internal/handlers/songs_handler.go`
- Build/Test: `server/internal/handlers/albums_handler.go`
- Build/Test: `web/`

- [ ] **Step 1: Build the backend**

Run:

```bash
cd server && go build ./...
```

Expected: successful Go build with no compile errors.

- [ ] **Step 2: Type-check the frontend**

Run:

```bash
cd web && bun run type-check
```

Expected: successful TypeScript verification with no regressions.

- [ ] **Step 3: Start the backend using the current local env path**

Run in its own terminal:

```bash
cd server && go run cmd/start_server/main.go
```

Expected startup evidence in logs:

```text
Loaded .env.dev
Connecting to postgres database:
Database connected successfully
S3 storage initialized
Starting Atoman Backend Server...
```

If you see `.env` loaded first, SQLite connection logs, or “falling back to local storage,” treat that as a failure against the documented DevOps design.

- [ ] **Step 4: Start the frontend dev server**

Run in a second terminal:

```bash
cd web && bun run dev
```

Expected: the Vite server starts successfully and exposes the local dev URL.

- [ ] **Step 5: Re-run the critical manual checks from upstream progress**

In the browser, verify these exact flows against the live app:

```text
- login succeeds
- successful login redirects to /feed
- RSS timeline loads
- music timeline displays MinIO-backed cover images
- album detail page can start playback
- player cover and audio both come from the MinIO/S3-backed path
- blog list page loads from PostgreSQL-backed data
- blog detail page loads from PostgreSQL-backed data
```

Expected: the same user-visible outcomes documented in `plan/dev_ops_progress.md` are still reproducible.

## Task 4: Verify the SQLite/uploads migration path still works when legacy artifacts are present

**Files:**
- Use: `server/migrate_db.go`
- Optional legacy inputs: `server/dev.sqlite`, `server/uploads/`
- Verify via: PostgreSQL in `docker-compose.dev.yml`

- [ ] **Step 1: Detect whether legacy migration inputs are available**

Run from the repository root:

```bash
python - <<'PY'
from pathlib import Path
sqlite = Path('server/dev.sqlite').exists()
uploads = Path('server/uploads').exists()
print(f'dev_sqlite={sqlite}')
print(f'uploads_dir={uploads}')
PY
```

Expected:

```text
- if both are false, record that migration rerun cannot be validated in this checkout
- if either is true, continue with the migration checks below
```

- [ ] **Step 2: Run the migration tool when legacy inputs exist**

Run:

```bash
cd server && go run migrate_db.go
```

Expected startup evidence in logs:

```text
Starting Migration: SQLite + Local Files -> PostgreSQL + MinIO
Connected to SQLite (dev.sqlite)
Connected to PostgreSQL
Connected to MinIO
--- Starting File Uploads ---
--- Starting Database Migration ---
--- Migration Complete! ---
```

- [ ] **Step 3: Verify PostgreSQL now contains migrated content**

Run from the repository root:

```bash
set -a && source .env.dev && set +a && \
  docker compose -f docker-compose.dev.yml exec -T postgres \
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  -c "select count(*) as users from \"Users\"; select count(*) as posts from posts; select count(*) as songs from \"Songs\";"
```

Expected: non-zero counts for whichever legacy datasets were present before migration.

- [ ] **Step 4: Verify migrated content is now served from the S3/MinIO URL pattern**

Use either the browser UI from Task 3 or direct database spot checks. At minimum, verify these exact expectations:

```text
- migrated avatar_url values no longer rely on /uploads/... when migrated
- migrated audio_url values no longer rely on /uploads/... when migrated
- migrated cover_url values no longer rely on /uploads/... when migrated
- music and blog pages render migrated content without referencing the old local path as the primary source
```

Expected: the old `/uploads/...` path is no longer the main runtime dependency after migration.

## Task 5: Revalidate the current production packaging path and make its limits explicit

**Files:**
- Inspect/Test: `Dockerfile`
- Inspect/Test: `docker-compose.prod.yml`
- Inspect/Test: `manage.sh`

- [ ] **Step 1: Confirm the production image still depends on prebuilt artifacts**

Inspect the files and confirm these responsibilities:

```text
- Dockerfile copies web/dist into nginx html root
- Dockerfile copies server/main into /app/main
- manage.sh builds server/main and web/dist before container startup
- docker-compose.prod.yml expects app image build plus server/.env.prod at runtime
```

Expected: the current production deployment path is accurately understood before attempting to automate it further.

- [ ] **Step 2: Build the backend production binary exactly as the current flow expects**

Run:

```bash
cd server && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o main ./cmd/start_server
```

Expected: `server/main` exists and is executable as the container entry binary.

- [ ] **Step 3: Build the frontend production bundle**

Run:

```bash
cd web && bun run build
```

Expected: `web/dist/` is regenerated successfully.

- [ ] **Step 4: Verify the runtime image can be built from the prebuilt artifacts**

Run from the repository root:

```bash
docker build -t atoman-prod-runtime-check -f Dockerfile .
```

Expected: the image builds successfully without missing-file errors for `web/dist` or `server/main`.

- [ ] **Step 5: Separate “current working template” from “future improvement” in working notes**

Record these conclusions explicitly:

```text
- current state: Compose + Dockerfile + prebuilt artifacts is the supported production template today
- current limitation: this is not yet the streamlined multi-stage build recommended in findings
- required follow-up if productized further: replace the prebuild-only runtime image path with a reproducible multi-stage build pipeline
```

Expected: nobody mistakes the present packaging path for a fully finished immutable deployment pipeline.

## Task 6: Close the documentation and artifact loop after verification

**Files:**
- Update if needed: `plan/dev_ops_progress.md`
- Update if needed: `plan/dev_ops_findings.md`
- Update if needed: environment-template docs if `.env.dev.example` is still expected but absent

- [ ] **Step 1: Record the exact verification date and outcomes**

Add a short dated note capturing:

```text
- whether Tasks 2-5 passed fully
- which checks were skipped because required local artifacts or secrets were absent
- whether the backend still preferred .env.dev and connected to postgres + s3 as intended
```

Expected: future sessions can tell the difference between historical 2026-05-12 success and the fresh revalidation run.

- [ ] **Step 2: Resolve the env-template naming mismatch explicitly**

If the repo still has `server/.env.example` but not a tracked repo-root `.env.dev.example`, choose one explicit source-of-truth path and document it in the plan/progress notes. Do not leave both names implied.

Expected: future setup instructions no longer reference a missing template file.

## Verification

After execution, verify all of the following:

1. `docker-compose.dev.yml` still represents an infrastructure-only local stack: PostgreSQL, db-init, MinIO, and MinIO setup.
2. `.env.dev` exists locally, includes the required keys, and sets `DATABASE_TYPE=postgres` plus `STORAGE_TYPE=s3`.
3. `docker compose --env-file .env.dev -f docker-compose.dev.yml up -d postgres db-init minio minio-setup` completes successfully.
4. `docker compose --env-file .env.dev -f docker-compose.dev.yml ps` shows `postgres` and `minio` healthy and the init containers exited successfully.
5. `curl -fsS http://127.0.0.1:9100/minio/health/live` succeeds.
6. `cd server && go build ./...` passes.
7. `cd web && bun run type-check` passes.
8. `cd server && go run cmd/start_server/main.go` loads `.env.dev`, connects to PostgreSQL, and initializes S3 storage.
9. `cd web && bun run dev` starts successfully.
10. Manual verification covers login, `/feed`, RSS timeline, music cover rendering, album playback, blog list, and blog detail.
11. If `server/dev.sqlite` and/or `server/uploads/` exist, `cd server && go run migrate_db.go` completes and migrated data is visible in PostgreSQL-backed runtime flows.
12. `cd server && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o main ./cmd/start_server` passes.
13. `cd web && bun run build` passes.
14. `docker build -t atoman-prod-runtime-check -f Dockerfile .` succeeds.
15. Any remaining mismatch around `.env.dev.example` versus `server/.env.example` is documented explicitly rather than left implicit.

## Commit

```bash
git add docs/superpowers/plans/2026-05-14-dev-ops-implementation.md
git commit -m "docs: add dev ops superpowers implementation plan"
```
