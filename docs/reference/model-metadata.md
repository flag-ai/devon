# Model Metadata

The `ModelMetadata` dataclass is the central data structure DEVON uses to
represent a model. It is returned by source plugins, displayed by CLI
commands, and persisted in the storage index.

## Location

```
devon.models.model_info.ModelMetadata
```

## Fields

| Field | Type | Description |
|-------|------|-------------|
| `source` | `str` | Source identifier (e.g., `"huggingface"`) |
| `model_id` | `str` | Full model ID (e.g., `"Qwen/Qwen2.5-32B"`) |
| `model_name` | `str` | Human-readable display name |
| `author` | `str` | Model author or organization |
| `total_size_bytes` | `int` | Total size of all model files in bytes |
| `file_count` | `int` | Number of files in the model repository |
| `parameter_count` | `Optional[int]` | Parameter count (in billions). `None` if unknown |
| `architecture` | `Optional[str]` | Model architecture (e.g., `llama`, `qwen`, `mistral`) |
| `format` | `List[str]` | Available formats: `gguf`, `safetensors`, `pytorch` |
| `quantization` | `Optional[str]` | Quantization type (e.g., `Q4_K_M`, `Q5_K_M`). `None` for full-precision |
| `tags` | `List[str]` | Tags from the model repository |
| `license` | `Optional[str]` | License identifier (e.g., `"apache-2.0"`, `"mit"`) |
| `downloads` | `int` | Total download count |
| `likes` | `int` | Total like count |
| `created_at` | `str` | ISO 8601 creation timestamp |
| `updated_at` | `str` | ISO 8601 last-modified timestamp |
| `web_url` | `str` | Browser-facing URL for the model page |
| `repo_url` | `str` | Repository URL used for cloning or API access |
| `extra` | `Dict[str, Any]` | Source-specific fields that do not fit the standard schema |

## Usage

Source plugins construct `ModelMetadata` instances in their `search()` and
`get_model_info()` methods:

```python
from devon.models.model_info import ModelMetadata

meta = ModelMetadata(
    source="huggingface",
    model_id="Qwen/Qwen2.5-32B-Instruct",
    model_name="Qwen2.5-32B-Instruct",
    author="Qwen",
    total_size_bytes=67_108_864_000,
    file_count=12,
    parameter_count=32,
    architecture="qwen",
    format=["safetensors"],
    quantization=None,
    tags=["text-generation", "conversational"],
    license="apache-2.0",
    downloads=150_000,
    likes=2_400,
    created_at="2025-01-15T10:00:00Z",
    updated_at="2025-02-01T08:30:00Z",
    web_url="https://huggingface.co/Qwen/Qwen2.5-32B-Instruct",
    repo_url="https://huggingface.co/Qwen/Qwen2.5-32B-Instruct",
    extra={},
)
```

## Notes

- The `license` field may come from HuggingFace as a list; the HuggingFace
  source plugin joins multiple values with `", "` before storing.
- The `extra` dict allows source plugins to carry additional data without
  modifying the shared dataclass.
- When serialized to the storage index, `ModelMetadata` is stored inside the
  `"metadata"` key of each index entry.
