# Source Plugins

DEVON uses a plugin system to support multiple model providers. Each source
is a self-contained class that knows how to search, fetch metadata, and
download models from one provider.

## ModelSource ABC

The abstract base class lives in `devon.sources.base` and defines five
methods that every source must implement:

| Method | Return Type | Description |
|--------|-------------|-------------|
| `name()` | `str` | Unique identifier for this source (e.g., `"huggingface"`) |
| `is_available()` | `bool` | Whether the source can be reached right now |
| `search(query, **filters)` | `list[ModelMetadata]` | Search for models matching the query and filters |
| `get_model_info(model_id)` | `ModelMetadata` | Fetch full metadata for a single model |
| `download_model(model_id, dest)` | `Path` | Download model files to `dest` and return the path |

All methods are expected to raise descriptive exceptions on failure so the
CLI layer can display useful error messages.

## SourceRegistry

The registry is defined in `devon.sources.registry` and uses a **class-level
dictionary** to track available sources:

```python
class SourceRegistry:
    _sources: dict[str, type[ModelSource]] = {}

    @classmethod
    def get_source(cls, name: str) -> ModelSource: ...

    @classmethod
    def list_available(cls) -> list[str]: ...

    @classmethod
    def list_all(cls) -> dict[str, type[ModelSource]]: ...
```

### @register_source Decorator

Sources register themselves at import time using the decorator:

```python
from devon.sources.registry import register_source
from devon.sources.base import ModelSource

@register_source
class MySource(ModelSource):
    def name(self) -> str:
        return "mysource"
    ...
```

The `sources/__init__.py` file imports all concrete source modules so that
registration happens automatically when the package loads.

## HuggingFace Implementation

The built-in HuggingFace source lives in `devon.sources.huggingface` and
depends on the `huggingface-hub` package (version ^0.32).

### Key Implementation Details

- **Search** uses `HfApi.list_models()` with **keyword arguments** directly
  (e.g., `author=`, `library=`, `search=`). The deprecated `ModelFilter`
  class was removed in HuggingFace Hub 2.0 and must not be used.

- **Card data** is a `ModelCardData` object, not a dictionary. Access fields
  with `getattr(model.card_data, "license", None)` rather than bracket
  notation.

- **Attribute names** are snake_case: `created_at`, `last_modified`,
  `card_data`, `model_id`.

- **License** may be returned as a list from the API. The implementation
  joins multiple values with `", "` before storing in `ModelMetadata`.

- **Downloads** use `snapshot_download()` from `huggingface_hub`. As of
  v0.32, resume behavior is the default and `resume_download=True` is no
  longer required.

## Adding a New Source

1. Create a new module under `devon/sources/` (e.g., `ollama.py`).
2. Define a class that inherits from `ModelSource` and implements all five
   abstract methods.
3. Decorate the class with `@register_source`.
4. Import the new module in `devon/sources/__init__.py`.
5. Add the source name to the `sources.enabled` list in the default
   configuration.

```python
# devon/sources/ollama.py
from devon.sources.base import ModelSource
from devon.sources.registry import register_source

@register_source
class OllamaSource(ModelSource):
    def name(self) -> str:
        return "ollama"

    def is_available(self) -> bool:
        ...
```

After these steps the new source will appear in `SourceRegistry.list_all()`
and can be selected with `--source ollama` on any CLI command.
