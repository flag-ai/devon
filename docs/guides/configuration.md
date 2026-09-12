# Configuration

DEVON reads its configuration from a YAML file. Every setting has a
sensible default so the config file is entirely optional.

## Config File Location

```
~/.config/devon/config.yaml
```

Create the file and its parent directory:

```bash
mkdir -p ~/.config/devon
$EDITOR ~/.config/devon/config.yaml
```

## Deep Merge

DEVON deep-merges your file over built-in defaults. Only specify keys
you want to change; omitted keys keep their defaults.

## Full Reference

### Storage

```yaml
storage:
  base_path: ~/.cache/devon/models   # Where models are saved
  max_size_gb: null                  # null = unlimited
```

Set `max_size_gb` to cap total cache size.

### Download

```yaml
download:
  resume: true              # Resume interrupted downloads
  verify_checksums: true    # Verify file integrity after download
```

### Sources

```yaml
sources:
  default: huggingface      # Source used when --source is omitted
  enabled:
    - huggingface
```

### Search

```yaml
search:
  default_limit: 20         # Number of results from devon search
  sort_by: downloads        # Sort order for search results
```

### Display

```yaml
display:
  color: true               # Enable Rich color output
```

Set to `false` when piping output to files or other tools.

## Common Examples

### Custom storage directory

```yaml
storage:
  base_path: /mnt/nvme/devon/models
```

### Cap storage at 500 GB

```yaml
storage:
  max_size_gb: 500
```

### Return 50 search results by default

```yaml
search:
  default_limit: 50
```

### Disable color output

```yaml
display:
  color: false
```

### Multiple overrides in one file

```yaml
storage:
  base_path: /data/llm-models
  max_size_gb: 1000
search:
  default_limit: 50
display:
  color: false
```
