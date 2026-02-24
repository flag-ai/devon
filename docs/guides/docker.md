# Docker Deployment

DEVON ships with a Dockerfile and docker-compose.yml for running the REST
API as a container. This is the recommended approach when you want KITT or
other remote clients to manage models over HTTP.

## Quick Start

```bash
docker compose up -d
curl http://localhost:8000/health
```

That's it. Models are stored in a Docker named volume (`devon-data`) by
default.

## Container Layout

The container maps everything under a single `/data` directory:

```
/data/
├── models/       # Downloaded model files
├── index.json    # Storage index
└── config.yaml   # Configuration (optional)
```

Mount a single host path to cover all three:

```bash
docker compose up -d  # uses named volume by default
# or
DEVON_DATA_PATH=/mnt/models docker compose up -d  # uses host path
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DEVON_PORT` | `8000` | Host port mapped to the container |
| `DEVON_DATA_PATH` | `devon-data` (named volume) | Host path or named volume for `/data` |
| `DEVON_API_KEY` | *(empty)* | Bearer token for API auth. Empty disables auth |
| `DEVON_FRAME_ANCESTORS` | *(empty)* | Origins allowed to embed Devon in an iframe (e.g., `https://kitt.example.com`). Falls back to `X-Frame-Options: DENY` when unset. |
| `HF_TOKEN` | *(empty)* | HuggingFace token for gated model access |

Set variables in a `.env` file next to `docker-compose.yml` or export them
in your shell.

### Example `.env`

```env
DEVON_PORT=9000
DEVON_DATA_PATH=/mnt/nvme/devon
DEVON_API_KEY=my-secret-token
HF_TOKEN=hf_abc123
```

## Building the Image

The Dockerfile uses a multi-stage build:

1. **Builder stage** — installs Poetry, resolves dependencies with `--extras api`,
   and creates an in-project virtualenv.
2. **Runtime stage** — copies only the virtualenv and source code into a clean
   `python:3.12-slim` image. Runs as a non-root `devon` user.

Build manually:

```bash
docker build -t devon .
```

Run manually (without Compose):

```bash
docker run -d \
  --name devon \
  -p 8000:8000 \
  -v /mnt/models:/data \
  -e DEVON_API_KEY=secret \
  -e HF_TOKEN=hf_abc123 \
  devon
```

## Health Check

The container includes a built-in Docker healthcheck that hits
`http://localhost:8000/health` every 30 seconds. Check container health:

```bash
docker inspect --format='{{.State.Health.Status}}' devon
```

## Using Your Existing Models

If you already have models downloaded on the host, mount their parent
directory as `/data`:

```bash
DEVON_DATA_PATH=/home/user/.cache/devon docker compose up -d
```

The container reads the existing `index.json` and serves models
immediately.

## Stopping

```bash
docker compose down       # stop and remove container (data volume preserved)
docker compose down -v    # stop and remove container AND named volume
```

## Production Tips

- **Single worker** — the default runs one Uvicorn worker. This avoids
  race conditions on the JSON index. Do not increase worker count without
  external write coordination.
- **Set `DEVON_API_KEY`** — in production, always set an API key to prevent
  unauthorized access.
- **Use host path mounts for persistence** — named volumes work but host
  paths make backups and migration easier.
- **Set `HF_TOKEN`** — required for downloading gated models (Llama, etc.).

---

## Further Reading

- [REST API Guide](rest-api.md) -- endpoint usage and examples
- [REST API Reference](../reference/rest-api.md) -- full request/response schemas
- [Configuration](configuration.md) -- YAML config options
