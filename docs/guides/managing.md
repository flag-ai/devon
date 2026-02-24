# Managing Local Models

Once you have downloaded models with DEVON, a set of commands lets you
list, inspect, and clean up your local collection.

## Listing Models

Show all locally downloaded models:

```bash
devon list
```

Filter by source:

```bash
devon list --source huggingface
```

## Model Details

Fetch detailed information about a model. DEVON queries HuggingFace for
live metadata and merges it with local storage information:

```bash
devon info Qwen/Qwen2.5-32B-Instruct
devon info meta-llama/Llama-3.3-70B-Instruct
```

## Storage Status

Get a summary of how much disk space your model collection uses:

```bash
devon status
```

The report shows total size, breakdown by source, and model count.

## Removing a Model

Delete a specific model from local storage:

```bash
devon remove Qwen/Qwen2.5-32B-Instruct
```

Skip the confirmation prompt with `--yes`:

```bash
devon remove Qwen/Qwen2.5-32B-Instruct --yes
```

## Cleaning Up

The `clean` command removes models based on usage criteria.

Remove unused models older than a threshold:

```bash
devon clean --unused --days 30
```

Preview what would be removed without deleting anything:

```bash
devon clean --dry-run
devon clean --all --dry-run
devon clean --unused --days 30 --dry-run
```

Remove all cached models:

```bash
devon clean --all
```

## Scanning for External Models

If you add models to the storage directory outside of Devon (custom
fine-tunes, manual copies, etc.), use `scan` to discover and register them:

```bash
devon scan
```

The scanner walks the directory tree, detects model weight files, and
infers metadata (format, architecture, quantization, parameter count)
from file extensions, filenames, and `config.json` if present.

Scan a different directory:

```bash
devon scan /data/custom-models
```

Preview what would be added without modifying the manifest:

```bash
devon scan --dry-run
```

Remove stale entries (models deleted outside Devon):

```bash
devon scan --reconcile
```

## Storage Directory Structure

DEVON organizes downloaded files under the configured base path:

```
~/.cache/devon/models/
├── manifest.json
├── huggingface/
│   ├── Qwen/Qwen2.5-32B-Instruct/
│   └── meta-llama/Llama-3.3-70B-Instruct/
└── local/
    └── my-custom-model/
```

- **huggingface/** (and other source directories) contain subdirectories
  for each `author/model-name` pair.
- **local/** contains models discovered by `devon scan` that were not
  downloaded through a known source.
- **manifest.json** tracks metadata for every model including source,
  model ID, download date, and file sizes. DEVON uses this manifest for
  fast lookups without scanning the filesystem.

You can change the storage base path in your
[configuration file](configuration.md).
