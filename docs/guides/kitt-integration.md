# KITT Integration

DEVON and KITT are companion tools. DEVON manages the models. KITT tests
them. Together they form a workflow for downloading models and running
inference benchmarks.

## Overview

[KITT](https://kirizan.github.io/kitt/) is an inference engine testing
suite that measures model performance across different serving backends.
DEVON's export command produces output in formats that KITT can consume
directly.

## Exporting for KITT

Generate a text file listing local model paths, one per line:

```bash
devon export --format kitt -o models.txt
```

The resulting `models.txt` file contains absolute paths to each
downloaded model directory, ready for KITT to read.

## Exporting as JSON

For programmatic consumption, export the full model index as JSON:

```bash
devon export --format json -o models.json
```

The JSON output includes model IDs, sources, file sizes, download dates,
and local paths.

## Using the Export with KITT

Pass the exported file to KITT's `--model-list` flag along with your
desired engine and test suite:

```bash
kitt run --model-list models.txt --engine vllm --suite standard
```

KITT reads each path from the file and runs the specified test suite
against every model in sequence.

## Full Workflow Example

A typical workflow from discovery to testing:

```bash
# 1. Search for candidate models
devon search "qwen instruct" --params 7b --format gguf

# 2. Download the ones you want
devon download Qwen/Qwen2.5-7B-Instruct

# 3. Export paths for KITT
devon export --format kitt -o models.txt

# 4. Run inference tests
kitt run --model-list models.txt --engine vllm --suite standard
```

This pattern scales to any number of models. Download as many as you
need, export once, and KITT tests them all.

### Working with Custom Models

If you have models from sources other than HuggingFace (custom fine-tunes,
converted weights, etc.), place them in the models directory and scan:

```bash
# Add your custom model to the storage directory
cp -r /path/to/my-custom-model ~/.cache/devon/models/local/my-custom-model

# Register it in the manifest
devon scan

# Export for KITT — custom models are included automatically
devon export --format kitt -o models.txt
```

## Portable Model Directory

The model manifest (`manifest.json`) lives inside the models directory
itself. This means you can copy or mount the entire directory on another
machine and KITT can consume it directly — no separate index file to
track.

## Remote Integration via REST API

When DEVON runs as a containerized API server, KITT (or any client) can
manage models over HTTP without requiring DEVON to be installed locally:

```bash
# Search for models
curl "http://devon-host:8000/api/v1/search?provider=qwen&limit=5"

# Download a model
curl -X POST http://devon-host:8000/api/v1/downloads \
  -H "Content-Type: application/json" \
  -d '{"model_id": "Qwen/Qwen2.5-7B-Instruct"}'

# Export model paths
curl -X POST http://devon-host:8000/api/v1/export \
  -H "Content-Type: application/json" \
  -d '{"format": "kitt"}'
```

See the [REST API guide](rest-api.md) and
[Docker Deployment guide](docker.md) for setup details.

## Further Reading

- [KITT documentation](https://kirizan.github.io/kitt/)
- [Searching for models](searching.md)
- [Downloading models](downloading.md)
- [Managing local models](managing.md)
- [REST API](rest-api.md)
- [Docker Deployment](docker.md)
