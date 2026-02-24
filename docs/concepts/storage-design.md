# Storage Design

DEVON stores downloaded models on disk and tracks them with a JSON index.
This page explains the directory layout, index format, and the design
decisions behind them.

## Directory Structure

Models are organized under the configured `storage.base_path` (default
`~/.cache/devon/models/`) using the pattern:

```
models/{source}/{author}/{model}/
```

For example:

```
~/.cache/devon/models/
├── manifest.json
└── huggingface/
    ├── Qwen/
    │   └── Qwen2.5-32B-Instruct/
    └── meta-llama/
        └── Llama-3-70B-Instruct/
```

This layout mirrors the `author/model` convention used by HuggingFace and
keeps models from different sources cleanly separated. The `manifest.json`
file lives inside the models directory, making the entire directory
self-contained and portable.

## Manifest

The manifest file lives at `{base_path}/manifest.json` -- inside the
models directory itself. It is a single JSON object where each key is a
model identifier in the format:

```
{source}::{model_id}
```

For example: `huggingface::Qwen/Qwen2.5-32B-Instruct`.

See the [Manifest reference](../reference/storage-index.md) for the
full entry schema.

### Migration from index.json

Prior to v1.2.0, the index file was stored at `{base_path}/../index.json`
(a sibling of the models directory). On first load, DEVON automatically
migrates the legacy `index.json` to `manifest.json` inside the models
directory and deletes the old file. No manual action is required.

## Why a Flat JSON File

- **Simplicity.** No external database to install, configure, or migrate.
  The index is a single portable file.
- **Transparency.** Users can inspect and even hand-edit the index with any
  text editor or JSON tool.
- **Portability.** Copying the models directory to another machine is
  enough to move the entire vault -- the manifest travels with the models.

The trade-off is that concurrent writes from multiple processes are not
safe without coordination. For CLI use this is not a practical concern.
When running the REST API server, DEVON defaults to a **single Uvicorn
worker** to avoid index corruption from concurrent writes.

## Atomic Operations

The `ModelStorage` class uses a **read-modify-write** pattern with file
locking:

1. Read the entire index into memory.
2. Apply the change (add, update, or remove an entry).
3. Write the full index back to disk.

This keeps the index consistent even if the process is interrupted, because
the write replaces the file atomically.

## Size Tracking

Each index entry records a `size_bytes` value that is calculated at
registration time by summing the sizes of all downloaded files. The
`get_total_size()` method sums across every entry to report overall vault
usage, which powers the `devon status` command.

## Last-Used Tracking

Every index entry includes a `last_used` timestamp that is updated each
time the model is accessed (for example, by `devon info` or
`devon export`). The `devon clean --unused --days N` command uses this
timestamp to identify models that have not been touched in `N` days,
making it easy to reclaim disk space without manually deciding which models
to keep.

## Directory Scanning

The `devon scan` command walks the model directory tree and registers
any models not already in the manifest. This is useful when:

- Models are copied into the directory manually
- Custom fine-tuned models are added outside the Devon workflow
- The manifest is lost or needs to be rebuilt from scratch

The scanner infers metadata from what it finds on disk:

- **Format** from file extensions (`.safetensors`, `.gguf`, `.bin`)
- **Architecture** from `config.json` (`model_type` field)
- **Quantization** from filenames (e.g., `Q4_K_M`, `fp16`)
- **Parameter count** from directory names (e.g., `7B`) or `config.json`
- **Size** from actual file sizes on disk

Models under a recognized source directory (e.g., `huggingface/`) are
assigned that source. All others are registered with source `local`.

The `--reconcile` flag also removes manifest entries whose model
directories no longer exist on disk.
