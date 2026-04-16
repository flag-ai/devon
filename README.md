# DEVON

> _"DEVON manages the models. KITT tests them."_

DEVON is the model catalog and placement layer of the [FLAG](https://github.com/flag-ai) platform. It's a Go web service (REST API + embedded React SPA) that discovers HuggingFace models, orchestrates on-host downloads via [BONNIE](https://github.com/flag-ai/bonnie) agents, and tells KITT where to find the weights when a benchmark run is queued.

## Architecture

```
Web UI (React/Vite, embedded)
    │
    ▼
REST API  (Chi v5, Bearer-auth'd)
    │
    ├─ Postgres (pgx + sqlc + golang-migrate)
    │       devon_models / devon_placements
    │       devon_bonnie_agents / devon_download_jobs
    │
    └─ BONNIE agents (one per GPU host)
            /api/v1/models/fetch     — download from HF to host
            /api/v1/models           — list staged models
            /api/v1/models/{id}      — delete staged model
```

Adapted from the [flag-commons](https://github.com/flag-ai/commons) reference shape — the same patterns run in [KARR](https://github.com/flag-ai/karr) and KITT.

## Quick start (Docker Compose)

```bash
cp .env.example .env
# edit .env: set POSTGRES_PASSWORD, optionally HF_TOKEN

docker compose up -d postgres
docker compose up devon
```

Then open <http://localhost:8080>. The UI calls `POST /api/v1/setup` on first load to generate an admin token — copy it somewhere safe; it won't be shown again.

Register your BONNIE agents in the **Agents** tab, then search and download.

## Local development

```bash
# Postgres
docker compose up -d postgres

# Backend
cp .env.example .env
go run ./cmd/devon serve

# Frontend (separate terminal, live-reload)
cd web && npm install && npm run dev
```

The Vite dev server (port 5173) proxies API calls to the Go server (port 8080).

## Configuration

| Env var                      | Default  | Purpose                                                                       |
| ---------------------------- | -------- | ----------------------------------------------------------------------------- |
| `DATABASE_URL`               | required | `postgres://…`                                                                |
| `LISTEN_ADDR`                | `:8080`  | HTTP listen address                                                           |
| `LOG_LEVEL`                  | `info`   | `debug`, `info`, `warn`, `error`                                              |
| `LOG_FORMAT`                 | `text`   | `text` or `json`                                                              |
| `DEVON_ADMIN_TOKEN`          | `""`     | Bearer token for `/api/v1/*`. Empty triggers the `/setup` flow on first load. |
| `HF_TOKEN`                   | `""`     | Optional HuggingFace token                                                    |
| `DEVON_CORS_ORIGINS`         | `""`     | Comma-separated allowed origins                                               |
| `DEVON_FRAME_ANCESTORS`      | `""`     | CSP `frame-ancestors` override for embedding                                  |

When `secrets.NewProvider` succeeds against OpenBao, secrets are sourced from `kv/devon`; otherwise the process falls back to env vars.

## API surface

All routes live under `/api/v1`, Bearer-authenticated (via `DEVON_ADMIN_TOKEN` or the token provisioned by `/setup`) except `/health`, `/ready`, `/setup`.

| Method | Path                               | Purpose                                   |
| ------ | ---------------------------------- | ----------------------------------------- |
| POST   | `/setup`                           | First-run admin token provisioning        |
| GET    | `/search`                          | HuggingFace model search                  |
| GET    | `/models`                          | List tracked models with placements       |
| GET    | `/models/{source}/{model_id}`      | Model detail                              |
| DELETE | `/models/{source}/{model_id}`      | Remove (fans out to every placement)      |
| POST   | `/models/download`                 | Queue a download job                      |
| GET    | `/downloads`                       | List jobs                                 |
| GET    | `/downloads/{id}`                  | Job detail                                |
| POST   | `/downloads/{id}/restart`          | Retry a finished/failed job               |
| POST   | `/models/ensure`                   | KITT: "ensure model X on host Y"          |
| GET    | `/bonnie-agents`                   | List registered BONNIE agents             |
| POST   | `/bonnie-agents`                   | Register an agent                         |
| DELETE | `/bonnie-agents/{id}`              | Deregister                                |
| POST   | `/scan`                            | Reconcile placements with a live agent    |
| POST   | `/export`                          | Emit JSON or TSV for KITT consumption     |
| GET    | `/config` / PUT                    | Non-secret config snapshot                |
| GET    | `/config/secrets` / PUT            | Masked secrets view + rotate              |

## Development workflow

```bash
make test         # go test -race ./...
make lint         # golangci-lint run ./...
make security     # gosec ./...
make sqlc         # regenerate internal/db/sqlc
make build        # go build -o devon ./cmd/devon
make build-web    # npm run build inside web/
make docker       # docker compose build
```

## Versioning

`VERSION` tracks the current release. Every PR bumps it — patch for cleanup, minor for new feature surface, major for a breaking change. The build pipeline threads `VERSION` / `COMMIT` / `BUILD_DATE` into `github.com/flag-ai/commons/version`.

## License

Apache 2.0 — see [LICENSE](LICENSE).
