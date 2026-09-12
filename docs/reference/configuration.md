# Configuration Schema

DEVON reads its configuration from a YAML file at:

```
~/.config/devon/config.yaml
```

If the file does not exist, DEVON uses built-in defaults for every setting.

## How Configuration Is Loaded

The `Settings` class in `devon.config.settings` handles configuration:

1. A `DEFAULT_CONFIG` dictionary defines every key and its default value.
2. If the YAML file exists, it is loaded and **deep-merged** over the
   defaults -- user values override defaults while unset keys keep their
   default.
3. Values are accessed with **dot-notation** through `settings.get()`:

```python
from devon.config.settings import Settings

settings = Settings()
base = settings.get("storage.base_path")   # "~/.cache/devon/models"
limit = settings.get("search.default_limit")  # 20
```

Convenience properties are also available:

| Property | Equivalent `get()` call |
|----------|------------------------|
| `settings.storage_path` | `settings.get("storage.base_path")` |
| `settings.default_source` | `settings.get("sources.default")` |
| `settings.search_limit` | `settings.get("search.default_limit")` |

## Full Schema

| Section | Key | Type | Default | Description |
|---------|-----|------|---------|-------------|
| `storage` | `base_path` | string | `~/.cache/devon/models` | Root directory for downloaded models |
| `storage` | `max_size_gb` | int or null | `null` | Maximum total storage in GB. `null` means unlimited |
| `download` | `resume` | bool | `true` | Resume interrupted downloads instead of restarting |
| `download` | `verify_checksums` | bool | `true` | Verify file checksums after download completes |
| `sources` | `default` | string | `huggingface` | Source plugin used when `--source` is not specified |
| `sources` | `enabled` | list[string] | `[huggingface]` | Source plugins that are active |
| `search` | `default_limit` | int | `20` | Number of results returned by `devon search` |
| `search` | `sort_by` | string | `downloads` | Default sort order: `downloads`, `likes`, or `last_modified` |
| `display` | `color` | bool | `true` | Enable colored output via Rich |

## Example File

```yaml
storage:
  base_path: ~/models
  max_size_gb: 100

download:
  resume: true
  verify_checksums: true

sources:
  default: huggingface
  enabled:
    - huggingface

search:
  default_limit: 50
  sort_by: likes

display:
  color: true
```

## Notes

- Paths containing `~` are expanded to the user's home directory at load
  time.
- Unknown keys are silently ignored, so older config files remain forward
  compatible.
- The deep-merge strategy means you only need to include the keys you want
  to change; everything else falls back to the default.
