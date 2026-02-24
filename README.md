# DEVON - Discovery Engine and Vault for Open Neural models

> *"DEVON manages the models. KITT tests them."*

CLI tool, REST API, and Web UI for discovering, downloading, and managing LLM models from HuggingFace and other sources.

[**Full Documentation**](https://kirizan.github.io/devon/) | [**CLI Reference**](https://kirizan.github.io/devon/reference/cli/)

## Features

- **Smart search** — filter by provider, size, parameters, format, task, license
- **Easy download** — by URL or model ID with automatic resume
- **Local vault** — organized storage with portable manifest and disk usage tracking
- **Directory scanning** — discover models added outside Devon (custom fine-tunes, manual copies) with automatic metadata inference
- **KITT integration** — export model paths for inference testing
- **Source plugins** — extensible architecture for model sources
- **YAML configuration** — deep-merged config with sensible defaults
- **Web UI** — browser-based dashboard for search, downloads, model management, and configuration
- **REST API** — FastAPI server for remote model management
- **Docker ready** — containerize with a single volume mount (Web UI included)

## Quick Start

```bash
# Install
poetry install
eval $(poetry env activate)

# Search for models
devon search --provider qwen --params 30b --format gguf

# Download by URL
devon download https://huggingface.co/Qwen/Qwen2.5-32B-Instruct

# List downloaded models
devon list

# Discover models added outside Devon
devon scan

# Export for KITT
devon export --format kitt -o models.txt
```

## Commands

| Command | Description |
|---|---|
| `devon search` | Search for models with filters ([filter guide](https://kirizan.github.io/devon/guides/searching/)) |
| `devon download` | Download a model by URL or ID |
| `devon list` | List downloaded models |
| `devon info` | Show model details |
| `devon status` | Storage usage summary |
| `devon scan` | Discover and register untracked models |
| `devon clean` | Remove old or unused models |
| `devon remove` | Delete a specific model |
| `devon export` | Export paths for KITT |
| `devon serve` | Start the REST API server |

### Search Filters

The `search` command supports these filters (combine freely, AND logic):

```bash
devon search "query"                          # text search
devon search --provider qwen                  # by author/org (-p)
devon search --params 30b                     # by parameter count (±20% tolerance)
devon search --size "<100gb"                  # by file size (<, <=, >, >=)
devon search --format gguf                    # by format (-f: gguf, safetensors, pytorch, onnx)
devon search --task text-generation           # by pipeline tag (-t)
devon search --license apache-2.0             # by license (-l)
devon search --limit 50                       # max results (default 20)
```

Filters also work inline: `devon search "qwen 30b gguf"` auto-extracts params and format.

See the [full filter guide](https://kirizan.github.io/devon/guides/searching/) for detailed examples and sample output.

## Documentation

| Section | Description |
|---|---|
| [Getting Started](https://kirizan.github.io/devon/getting-started/) | Installation and first model tutorial |
| [Guides](https://kirizan.github.io/devon/guides/) | Searching, downloading, managing, configuration |
| [Reference](https://kirizan.github.io/devon/reference/) | CLI reference, config schema, data models |
| [Concepts](https://kirizan.github.io/devon/concepts/) | Architecture, source plugins, storage design |

## Configuration

Config file at `~/.config/devon/config.yaml`. Only override what you need:

```yaml
storage:
  base_path: /mnt/data/models
  max_size_gb: 500
```

See the [full configuration guide](https://kirizan.github.io/devon/guides/configuration/) for all options.

## REST API

Install the API extras, then start the server:

```bash
poetry install --extras api
devon serve                       # http://127.0.0.1:8000
devon serve --host 0.0.0.0 --port 9000
```

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check (no auth) |
| GET | `/api/v1/setup/status` | Check if first-run setup is needed (no auth) |
| POST | `/api/v1/setup` | Generate API key during first-run setup (no auth) |
| GET | `/api/v1/search` | Search remote models |
| GET | `/api/v1/models` | List local models |
| GET | `/api/v1/models/{source}/{model_id}` | Model info (local + remote) |
| DELETE | `/api/v1/models/{source}/{model_id}` | Remove a model |
| POST | `/api/v1/downloads` | Start a download (returns 200 cached or 202 accepted) |
| GET | `/api/v1/downloads` | List all download jobs |
| GET | `/api/v1/downloads/{job_id}` | Get download job status |
| POST | `/api/v1/downloads/{job_id}/restart` | Restart a failed download |
| GET | `/api/v1/status` | Storage stats |
| POST | `/api/v1/scan` | Scan for untracked models |
| POST | `/api/v1/clean` | Clean unused models |
| POST | `/api/v1/export` | Export model list |
| GET | `/api/v1/config` | Get configuration (secrets masked) |
| PUT | `/api/v1/config` | Update configuration |
| GET | `/api/v1/config/setup-status` | First-run setup status |
| PUT | `/api/v1/config/secrets` | Set HF token / API key (write-only) |

### Authentication

DEVON uses three-tier authentication on `/api/v1/*` endpoints (checked in order):

| Tier | Source | Behavior |
|------|--------|----------|
| 1 | `DEVON_API_KEY` env var | Requires `Authorization: Bearer <key>` header. Set to `disable` to skip auth. |
| 2 | Config file (`secrets.api_key`) | Same bearer token auth, key stored via first-run setup or `PUT /api/v1/config/secrets` |
| 3 | Neither set | Returns **503** `DEVON_SETUP_REQUIRED` — triggers the Web UI first-run setup flow |

On first launch with no API key configured, the Web UI walks you through setup at `POST /api/v1/setup`, which generates and persists a key automatically.

```bash
# Production — require a bearer token
DEVON_API_KEY=mysecretkey devon serve
curl -H "Authorization: Bearer mysecretkey" http://localhost:8000/api/v1/models

# Local development — explicitly disable auth
DEVON_API_KEY=disable devon serve
curl http://localhost:8000/api/v1/models
```

> **Security note:** Environment variables are visible in `/proc/<pid>/environ` and in `ps e` output. On shared hosts, prefer injecting `DEVON_API_KEY` via a secrets manager or a file-based mechanism rather than a shell `export`.

### Quick Examples

```bash
# Health check
curl http://localhost:8000/health

# Search for models
curl "http://localhost:8000/api/v1/search?provider=qwen&limit=3"

# List local models
curl http://localhost:8000/api/v1/models

# Start a download (returns 202 Accepted with job ID)
curl -X POST http://localhost:8000/api/v1/downloads \
  -H "Content-Type: application/json" \
  -d '{"model_id": "Qwen/Qwen2.5-1.5B"}'

# Check download progress
curl http://localhost:8000/api/v1/downloads

# Storage status
curl http://localhost:8000/api/v1/status
```

## Web UI

Devon includes a browser-based dashboard for managing models without the CLI. When `devon serve` starts, the Web UI is available at the root URL (e.g. `http://localhost:8000/`).

**Pages:**

| Page | Description |
|------|-------------|
| Dashboard | Storage stats, recent models, quick search |
| Search | Full search with all filter controls |
| Models | Browse, inspect, and delete local models |
| Downloads | Start downloads, track progress, restart failed jobs |
| Settings | Configure all settings and secrets from the browser |

On first launch (no API key configured), the Web UI presents a setup flow that generates and stores an API key automatically. Devon works immediately with defaults — configuration is recommended but not required.

### Building the frontend

The Web UI is built from `frontend/` (React + Vite + TypeScript + Tailwind CSS). The Docker image builds it automatically. For local development:

```bash
# Build the UI into src/devon/ui/static/
make build-ui

# Or run the Vite dev server with API proxy
make dev-ui    # UI at http://localhost:5173, proxies /api to :8000
```

## Docker

### Build and run

```bash
docker compose up -d
curl http://localhost:8000/health   # API health check
# Open http://localhost:8000/ in a browser for the Web UI
```

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DEVON_PORT` | `8000` | Host port mapping |
| `DEVON_DATA_PATH` | `devon-data` (named volume) | Host path for model storage |
| `DEVON_API_KEY` | *(empty — 503 until set)* | Bearer token for API endpoints (`disable` to skip auth) |
| `HF_TOKEN` | *(empty)* | HuggingFace token for gated models |

Mount your existing models directory:

```bash
DEVON_DATA_PATH=/mnt/models docker compose up -d
```

The container stores models (and the manifest) at `/data/models/` and config at `/data/config.yaml`. A single `-v /your/path:/data` covers everything.

**Note:** The default configuration runs a single uvicorn worker to avoid race conditions on the JSON index file.

## Related Projects

- [KITT](https://github.com/kirizan/kitt) — LLM inference testing suite ([docs](https://kirizan.github.io/kitt/))

## License

Apache 2.0
