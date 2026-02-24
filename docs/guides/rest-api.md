# REST API

DEVON includes a FastAPI-based REST API server that exposes every core
capability over HTTP. Use it when you need to manage models remotely --
for example, from KITT or a CI pipeline -- without installing DEVON on
the client machine.

## Prerequisites

Install the optional API dependencies:

```bash
poetry install --extras api
```

This adds FastAPI and Uvicorn. The standard CLI continues to work without
these extras.

## Starting the Server

```bash
devon serve                             # http://127.0.0.1:8000
devon serve --host 0.0.0.0 --port 9000 # bind to all interfaces
devon serve --reload                    # auto-reload on code changes (dev)
```

!!! tip
    For containerized deployments, see the
    [Docker Deployment](docker.md) guide instead of running `devon serve`
    directly.

## Authentication

DEVON uses three-tier authentication on all `/api/v1/*` endpoints (checked
in order):

1. **`DEVON_API_KEY` env var** — if set, requires `Authorization: Bearer <key>`.
   Set to `disable` to explicitly skip auth.
2. **Config file** (`secrets.api_key`) — same bearer auth, key stored via
   first-run setup or `PUT /api/v1/config/secrets`.
3. **Neither set** — returns **503** `DEVON_SETUP_REQUIRED`, triggering
   the Web UI first-run setup flow.

The `/health` and `/api/v1/setup/*` endpoints are always unauthenticated.

```bash
DEVON_API_KEY=my-secret devon serve
```

Clients pass the token in the `Authorization` header:

```bash
curl -H "Authorization: Bearer my-secret" http://localhost:8000/api/v1/models
```

## Endpoint Overview

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check (no auth) |
| GET | `/api/v1/setup/status` | Check if first-run setup is needed (no auth) |
| POST | `/api/v1/setup` | Generate API key during first-run setup (no auth) |
| GET | `/api/v1/search` | Search remote models |
| GET | `/api/v1/models` | List local models |
| GET | `/api/v1/models/{source}/{model_id}` | Model info (local + remote) |
| DELETE | `/api/v1/models/{source}/{model_id}` | Remove a model |
| POST | `/api/v1/downloads` | Start a download (200 cached or 202 accepted) |
| GET | `/api/v1/downloads` | List all download jobs |
| GET | `/api/v1/downloads/{job_id}` | Get download job status |
| POST | `/api/v1/downloads/{job_id}/restart` | Restart a failed download |
| GET | `/api/v1/status` | Storage stats |
| POST | `/api/v1/clean` | Clean unused models |
| POST | `/api/v1/export` | Export model list |
| GET | `/api/v1/config` | Get configuration (secrets masked) |
| PUT | `/api/v1/config` | Update configuration |
| GET | `/api/v1/config/setup-status` | First-run setup status |
| PUT | `/api/v1/config/secrets` | Set HF token / API key (write-only) |

See the [REST API Reference](../reference/rest-api.md) for full request and
response schemas.

---

## Usage Examples

### Health check

```bash
curl http://localhost:8000/health
```

```json
{"status": "ok", "version": "1.0.0"}
```

### Search for models

Query parameters mirror the CLI's `--provider`, `--params`, `--size`,
`--format`, `--task`, `--license`, and `--limit` flags:

```bash
curl "http://localhost:8000/api/v1/search?provider=qwen&params=7b&limit=3"
```

### List local models

```bash
curl http://localhost:8000/api/v1/models
```

Filter by source:

```bash
curl "http://localhost:8000/api/v1/models?source=huggingface"
```

### Get model info

Returns both local storage info and remote metadata:

```bash
curl http://localhost:8000/api/v1/models/huggingface/Qwen/Qwen2.5-7B-Instruct
```

### Start a download

Downloads are asynchronous. The `POST` validates the model exists (fast API
call), then launches a background job and returns immediately:

```bash
curl -X POST http://localhost:8000/api/v1/downloads \
  -H "Content-Type: application/json" \
  -d '{"model_id": "Qwen/Qwen2.5-7B-Instruct"}'
```

**Response codes:**

- **200** — model already downloaded (and `force` is false). Body contains
  `{"cached": {...}}` with the existing download info.
- **202** — download job created. Body contains `{"job": {...}}` with the
  job ID and status.
- **404** — model not found on the source. No job is created.

Request body fields:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model_id` | string | *(required)* | Model identifier |
| `source` | string | `"huggingface"` | Source plugin name |
| `force` | bool | `false` | Re-download even if already present |
| `include_patterns` | list[string] | `null` | Glob patterns to filter files |

### List download jobs

```bash
curl http://localhost:8000/api/v1/downloads
```

Returns all tracked jobs (downloading, completed, failed) sorted newest
first. Each job includes `id`, `model_id`, `source`, `status`, timestamps,
and `result` (for completed jobs) or `error` (for failed jobs).

### Get download job status

```bash
curl http://localhost:8000/api/v1/downloads/{job_id}
```

### Restart a failed download

```bash
curl -X POST http://localhost:8000/api/v1/downloads/{job_id}/restart
```

Creates a new job with `force=True` for the same model. Only works on
jobs with status `failed`.

!!! note "In-memory job tracking"
    Download jobs are stored in memory. If the server restarts, job history
    is lost. Completed downloads remain in the storage index regardless.

### Delete a model

```bash
curl -X DELETE http://localhost:8000/api/v1/models/huggingface/Qwen/Qwen2.5-7B-Instruct
```

### Storage status

```bash
curl http://localhost:8000/api/v1/status
```

### Clean unused models

```bash
curl -X POST http://localhost:8000/api/v1/clean \
  -H "Content-Type: application/json" \
  -d '{"unused": true, "days": 30}'
```

Preview without deleting:

```bash
curl -X POST http://localhost:8000/api/v1/clean \
  -H "Content-Type: application/json" \
  -d '{"unused": true, "days": 30, "dry_run": true}'
```

### Export model list

```bash
curl -X POST http://localhost:8000/api/v1/export \
  -H "Content-Type: application/json" \
  -d '{"format": "kitt"}'
```

---

## Environment Variables

The API server respects these environment variables:

| Variable | Description |
|----------|-------------|
| `DEVON_API_KEY` | Bearer token for authentication (`disable` = no auth, empty = use config or trigger setup) |
| `DEVON_STORAGE_PATH` | Override the model storage directory |
| `DEVON_CONFIG_PATH` | Override the config file path |
| `DEVON_FRAME_ANCESTORS` | Space-separated origins allowed to embed Devon in an iframe (e.g., `https://kitt.example.com`). Sends `Content-Security-Policy: frame-ancestors` instead of `X-Frame-Options: DENY`. Invalid origins are silently rejected; falls back to DENY if all are invalid or unset. |
| `HF_TOKEN` | HuggingFace token for gated model access |

---

## Limitations

- **Single worker** — the default runs one Uvicorn worker to avoid race
  conditions on JSON index writes. Do not increase worker count without
  external write coordination.
- **In-memory job tracking** — download job history is lost on server
  restart. Completed downloads persist in the storage index.

---

## Further Reading

- [REST API Reference](../reference/rest-api.md) -- full request/response schemas
- [Docker Deployment](docker.md) -- containerized deployment
- [KITT Integration](kitt-integration.md) -- using the API with KITT
