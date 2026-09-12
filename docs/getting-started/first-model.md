# Tutorial: Your First Model

This tutorial walks you through the core DEVON workflow: search for a model, inspect
it, download it, and export the path for KITT.

!!! info "Prerequisites"
    Complete the [Installation](installation.md) guide first. You should be able to
    run `devon --version` successfully.

---

## 1. Search for Models

Use a free-text query or structured filters to find models.

```bash
devon search "llama 3"
```

DEVON queries HuggingFace and displays results in a table with name, provider,
parameter count, size, and download count. Add filters to narrow results:

```bash
devon search --provider qwen --params 30b
```

```bash
devon search --provider meta-llama --params 70b --size "<150gb" --license apache-2.0
```

!!! tip
    Use `--limit` to control how many results are shown. The default is 20.

---

## 2. Get Model Details

Pull up full metadata for a model before downloading:

```bash
devon info Qwen/Qwen2.5-32B-Instruct
```

This shows parameter count, total size, file list, license, tags, and dates.

---

## 3. Download a Model

Download by model ID:

```bash
devon download Qwen/Qwen2.5-32B-Instruct
```

Or by full HuggingFace URL -- DEVON auto-detects the source:

```bash
devon download https://huggingface.co/Qwen/Qwen2.5-32B-Instruct
```

!!! note
    Downloads resume automatically if interrupted. Run the same command again and
    DEVON picks up where it left off.

To force a fresh re-download:

```bash
devon download Qwen/Qwen2.5-32B-Instruct --force
```

---

## 4. Verify the Download

List all models in your local vault:

```bash
devon list
```

You should see a table with your newly downloaded model, its source, size, and date.

---

## 5. Check Storage Usage

```bash
devon status
```

This shows total model count, storage consumed, and vault location. If you have
configured a `max_size_gb` limit, remaining capacity is shown as well.

---

## 6. Export for KITT

Export downloaded model paths in a format KITT can consume:

```bash
devon export --format kitt -o models.txt
```

Then pass the file to KITT:

```bash
kitt run --model-list models.txt --engine vllm --suite standard
```

!!! tip
    Use `--format json` for structured output suitable for scripts or other tooling.

---

## Recap

| Step | Command | Purpose |
|---|---|---|
| Search | `devon search` | Find models on HuggingFace |
| Inspect | `devon info` | View detailed model metadata |
| Download | `devon download` | Pull model weights to local storage |
| List | `devon list` | See what is in your vault |
| Status | `devon status` | Check disk usage |
| Export | `devon export` | Produce paths for KITT or other tools |

---

!!! tip "Already have models on disk?"
    If you have models from other sources (custom fine-tunes, manual downloads),
    copy them into the storage directory and run `devon scan` to register them.
    See [Managing Models](../guides/managing.md) for details.

## Next Steps

- Advanced search filters: [Searching Models](../guides/searching.md)
- Vault management: [Managing Models](../guides/managing.md)
- Custom configuration: [Configuration](../guides/configuration.md)
