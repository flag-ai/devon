# REST API Reference

Technical reference for every endpoint in the DEVON REST API. For usage
examples and getting-started instructions, see the
[REST API guide](../guides/rest-api.md).

## Base URL

```
http://{host}:{port}
```

Default: `http://127.0.0.1:8000`

## Authentication

When the `DEVON_API_KEY` environment variable is set, all `/api/v1/*`
endpoints require a bearer token:

```
Authorization: Bearer {token}
```

The `/health` endpoint never requires authentication.

Unauthorized requests receive:

```json
{"detail": "Invalid or missing API key"}
```

**Status code:** `401 Unauthorized`

---

## Endpoints

### GET /health

Health check. Always unauthenticated.

**Response** `200`

```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

---

### GET /api/v1/search

Search remote model sources.

**Query Parameters**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `query` | string | `null` | Free-text search query |
| `source` | string | `"huggingface"` | Source plugin to query |
| `provider` | string | `null` | Filter by author/organization |
| `params` | string | `null` | Parameter count (e.g., `"7b"`, `"30b"`) |
| `size` | string | `null` | Size constraint (e.g., `"<100gb"`) |
| `format` | string | `null` | Model format (`gguf`, `safetensors`, `pytorch`) |
| `task` | string | `null` | Pipeline tag (e.g., `"text-generation"`) |
| `license` | string | `null` | License identifier (e.g., `"apache-2.0"`) |
| `limit` | int | `20` | Maximum results |

**Response** `200`

```json
{
  "query": "llama",
  "source": "huggingface",
  "count": 3,
  "results": [
    {
      "source": "huggingface",
      "model_id": "meta-llama/Llama-3.1-8B-Instruct",
      "model_name": "Llama-3.1-8B-Instruct",
      "author": "meta-llama",
      "total_size_bytes": 16106127360,
      "file_count": 5,
      "parameter_count": 8,
      "architecture": "llama",
      "format": ["safetensors"],
      "quantization": null,
      "tags": ["text-generation"],
      "license": "llama3.1",
      "downloads": 985000,
      "likes": 2400,
      "created_at": "2024-07-23T10:00:00+00:00",
      "updated_at": "2024-12-01T08:30:00+00:00",
      "web_url": "https://huggingface.co/meta-llama/Llama-3.1-8B-Instruct",
      "repo_url": "https://huggingface.co/meta-llama/Llama-3.1-8B-Instruct/tree/main"
    }
  ]
}
```

---

### GET /api/v1/models

List locally downloaded models.

**Query Parameters**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `source` | string | `null` | Filter by source name |

**Response** `200`

```json
{
  "count": 1,
  "models": [
    {
      "source": "huggingface",
      "model_id": "Qwen/Qwen2.5-7B-Instruct",
      "path": "/data/models/huggingface/Qwen/Qwen2.5-7B-Instruct",
      "size_bytes": 14495514624,
      "downloaded_at": "2025-02-12T14:30:00.123456",
      "last_used": null,
      "files": ["model.safetensors", "config.json", "tokenizer.json"],
      "metadata": {}
    }
  ]
}
```

---

### GET /api/v1/models/{source}/{model_id}

Get info for a specific model. Returns both local storage data and remote
metadata when available.

**Path Parameters**

| Parameter | Description |
|-----------|-------------|
| `source` | Source name (e.g., `huggingface`) |
| `model_id` | Model identifier (e.g., `Qwen/Qwen2.5-7B-Instruct`) |

**Response** `200`

```json
{
  "local": {
    "source": "huggingface",
    "model_id": "Qwen/Qwen2.5-7B-Instruct",
    "path": "/data/models/huggingface/Qwen/Qwen2.5-7B-Instruct",
    "size_bytes": 14495514624,
    "downloaded_at": "2025-02-12T14:30:00.123456",
    "last_used": null,
    "files": [],
    "metadata": {}
  },
  "remote": {
    "source": "huggingface",
    "model_id": "Qwen/Qwen2.5-7B-Instruct",
    "model_name": "Qwen2.5-7B-Instruct",
    "author": "Qwen",
    "total_size_bytes": 14495514624,
    "file_count": 8,
    "parameter_count": 7,
    "architecture": "qwen",
    "format": ["safetensors"],
    "quantization": null,
    "tags": [],
    "license": "apache-2.0",
    "downloads": 892000,
    "likes": 1800,
    "created_at": "",
    "updated_at": "",
    "web_url": "https://huggingface.co/Qwen/Qwen2.5-7B-Instruct",
    "repo_url": "https://huggingface.co/Qwen/Qwen2.5-7B-Instruct/tree/main"
  }
}
```

**Response** `404` — model not found locally or remotely.

---

### DELETE /api/v1/models/{source}/{model_id}

Remove a model from local storage.

**Path Parameters**

| Parameter | Description |
|-----------|-------------|
| `source` | Source name |
| `model_id` | Model identifier |

**Response** `200`

```json
{
  "deleted": true,
  "model_id": "Qwen/Qwen2.5-7B-Instruct",
  "source": "huggingface"
}
```

**Response** `404` — model not found.

---

### POST /api/v1/downloads

Download a model. Runs synchronously — set long client timeouts.

**Request Body**

```json
{
  "model_id": "Qwen/Qwen2.5-7B-Instruct",
  "source": "huggingface",
  "force": false,
  "include_patterns": null
}
```

| Field | Type | Default | Required | Description |
|-------|------|---------|----------|-------------|
| `model_id` | string | | Yes | Model identifier |
| `source` | string | `"huggingface"` | No | Source plugin name |
| `force` | bool | `false` | No | Re-download even if present |
| `include_patterns` | list[string] | `null` | No | Glob patterns to filter files (e.g., `["*Q4_K_M*"]`) |

**Response** `200`

```json
{
  "model_id": "Qwen/Qwen2.5-7B-Instruct",
  "source": "huggingface",
  "path": "/data/models/huggingface/Qwen/Qwen2.5-7B-Instruct",
  "files": ["model.safetensors", "config.json"],
  "size_bytes": 14495514624
}
```

**Response** `404` — model not found on remote source.

**Response** `500` — download failed.

---

### GET /api/v1/status

Get storage statistics.

**Response** `200`

```json
{
  "model_count": 3,
  "total_size_bytes": 94371840000,
  "storage_path": "/data/models",
  "sources": {
    "huggingface": {
      "count": 3,
      "size_bytes": 94371840000
    }
  }
}
```

---

### POST /api/v1/clean

Clean unused or all models.

**Request Body**

```json
{
  "unused": true,
  "days": 30,
  "all": false,
  "dry_run": false
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `unused` | bool | `false` | Remove models not used within `days` |
| `days` | int | `30` | Unused threshold in days |
| `all` | bool | `false` | Remove all models |
| `dry_run` | bool | `false` | List what would be removed without deleting |

**Response** `200`

```json
{
  "removed": 2,
  "freed_bytes": 28991029248,
  "dry_run": false,
  "models": [
    "huggingface/Qwen/Qwen2.5-7B-Instruct",
    "huggingface/meta-llama/Llama-3.2-3B-Instruct"
  ]
}
```

---

### POST /api/v1/export

Export model list.

**Request Body**

```json
{
  "format": "kitt"
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `format` | string | `"kitt"` | Export format: `"kitt"` (path list) or `"json"` (full details) |

**Response** `200` (kitt format)

```json
{
  "format": "kitt",
  "count": 2,
  "content": [
    "/data/models/huggingface/Qwen/Qwen2.5-7B-Instruct",
    "/data/models/huggingface/meta-llama/Llama-3.1-8B-Instruct"
  ]
}
```

**Response** `200` (json format)

```json
{
  "format": "json",
  "count": 1,
  "content": [
    {
      "source": "huggingface",
      "model_id": "Qwen/Qwen2.5-7B-Instruct",
      "path": "/data/models/huggingface/Qwen/Qwen2.5-7B-Instruct",
      "size_bytes": 14495514624,
      "downloaded_at": "2025-02-12T14:30:00.123456",
      "files": ["model.safetensors", "config.json"]
    }
  ]
}
```

---

### POST /api/v1/scan

Scan the model directory to discover untracked models and optionally
remove stale entries.

**Request Body**

```json
{
  "reconcile": false,
  "dry_run": false,
  "path": null
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `reconcile` | bool | `false` | Also remove entries whose files no longer exist on disk |
| `dry_run` | bool | `false` | Preview changes without modifying the manifest |
| `path` | string | `null` | Directory to scan (defaults to configured storage path) |

**Response** `200`

```json
{
  "added": 2,
  "existing": 5,
  "stale": 1,
  "removed": 0,
  "models": [
    {
      "model_id": "my-models/custom-llama-7B",
      "source": "local",
      "size_bytes": 4123456789,
      "status": "new"
    },
    {
      "model_id": "Qwen/Qwen2.5-7B-Instruct",
      "source": "huggingface",
      "size_bytes": 14495514624,
      "status": "new"
    },
    {
      "model_id": "old/deleted-model",
      "source": "huggingface",
      "size_bytes": 8000000000,
      "status": "stale"
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `added` | int | Number of new models registered |
| `existing` | int | Number of models already tracked |
| `stale` | int | Number of entries with missing files |
| `removed` | int | Number of stale entries removed (only when `reconcile` is true) |
| `models` | list | Per-model results with `status`: `new`, `stale`, or `removed` |

---

## Error Responses

All error responses follow this format:

```json
{
  "detail": "Error description here"
}
```

| Status Code | Meaning |
|-------------|---------|
| `400` | Bad request (invalid source name, malformed input) |
| `401` | Unauthorized (missing or invalid API key) |
| `404` | Resource not found |
| `422` | Validation error (invalid request body) |
| `500` | Internal server error (download failure, etc.) |
