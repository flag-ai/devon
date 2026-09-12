# CLI Reference

DEVON exposes ten commands through the `devon` entry point:

| Command | Purpose |
|---------|---------|
| `devon search` | Search for models on a remote source |
| `devon download` | Download a model by ID or URL |
| `devon list` | List locally downloaded models |
| `devon info` | Show detailed metadata for a model |
| `devon status` | Display vault disk usage and statistics |
| `devon scan` | Discover and register untracked models |
| `devon clean` | Remove unused or all cached models |
| `devon export` | Export model paths for KITT or JSON |
| `devon remove` | Delete a specific model from the vault |
| `devon serve` | Start the REST API server |

All commands support `--help` for inline usage information.

---

::: mkdocs-click
    :module: devon.cli.main
    :command: cli
    :prog_name: devon
    :depth: 2
    :style: table
