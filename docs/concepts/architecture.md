# Architecture

DEVON is organized as a set of focused packages under `src/devon/`. Each
package owns one responsibility, and the CLI layer ties them together.

## Project Structure

```
src/devon/
├── cli/           # Click commands and the main entry point group
├── api/           # FastAPI REST API (app factory, routers, schemas)
│   └── routers/   # Endpoint handlers (health, search, models, download, storage)
├── sources/       # Model source plugins (ABC, registry, implementations)
├── storage/       # Local model storage and JSON index management
├── search/        # Query parsing and filter logic
├── download/      # Download orchestration with resume support
├── models/        # ModelMetadata dataclass
├── config/        # Settings class with YAML loading and deep merge
└── utils/         # URL parser, size parser, format detection helpers
```

### cli

The entry point is `devon.cli.main:cli`, a Click group that collects
sub-commands from individual modules (`search_cmd.py`, `download_cmd.py`,
`serve_cmd.py`, etc.). Each command module is self-contained: it parses
arguments, calls into the appropriate package, and renders output with Rich.

### api

A FastAPI application that exposes DEVON's capabilities over HTTP. The
`create_app()` factory in `api/app.py` initializes shared resources
(`Settings`, `ModelStorage`) via a lifespan context manager and registers
routers for health, search, models, downloads, and storage operations.
Routers call the same internal classes the CLI uses -- no logic is
duplicated. See the [REST API guide](../guides/rest-api.md) for usage.

### sources

Defines how DEVON talks to external model providers. The package contains an
abstract base class (`ModelSource`), a registry, and concrete
implementations. See [Source Plugins](source-plugins.md) for details.

### storage

Manages the on-disk model vault and its JSON index. See
[Storage Design](storage-design.md) for the file layout and index format.

### search

Provides query parsing that translates CLI filter flags (provider, parameter
count, size, format, task, license) into source-specific API calls.

### download

Handles downloading model files from a source, including progress tracking
and automatic resume of interrupted transfers.

### models

Contains the `ModelMetadata` dataclass -- the standard representation of a
model across every part of the system.

### config

Loads `~/.config/devon/config.yaml`, deep-merges it over built-in defaults,
and exposes values through dot-notation access.

### utils

Small standalone helpers: URL detection and parsing (`URLParser`), byte-size
formatting (`format_bytes`, `format_number`), and format detection.

## Plugin Registry Pattern

DEVON uses a **class-level dictionary registry** shared by source plugins:

1. An abstract base class (`ModelSource`) defines the interface.
2. A `SourceRegistry` class holds a `dict[str, type[ModelSource]]` as a
   class attribute.
3. A `@register_source` decorator adds a source class to the registry at
   import time.
4. The `sources/__init__.py` module imports all concrete sources so that
   registration happens automatically when the package is first loaded.

This pattern keeps source implementations decoupled from the rest of the
codebase and makes it straightforward to add new providers.

## Shared Patterns with KITT

DEVON and KITT share several conventions:

- **Poetry** for dependency management and virtual environments
- **Click** for CLI parsing
- **Rich** for terminal output (tables, panels, spinners)
- **Python 3.10+** minimum version
- **Plugin registries** with ABC + decorator + class-level dict

This consistency means contributors familiar with one project can navigate
the other with minimal ramp-up.

## Data Flow

A typical operation follows this path, regardless of whether it originates
from the CLI or the REST API:

```
CLI command / API request
  → Source plugin (search / get_model_info / download_model)
    → ModelMetadata dataclass
      → ModelStorage (register, index write)
        → JSON index on disk
```

1. The user invokes a CLI command (e.g., `devon download Qwen/Qwen2.5-32B`)
   or sends an HTTP request to the API.
2. The CLI or API router resolves the source plugin and calls its method.
3. The source plugin returns a `ModelMetadata` instance.
4. `ModelStorage` writes the model files to disk and registers the entry in
   the JSON index.
5. The CLI renders a summary with Rich, or the API returns a JSON response.
