from dataclasses import dataclass, field
from typing import List, Dict, Any, Optional


@dataclass
class ModelMetadata:
    """Universal model metadata across sources."""

    # Identity
    source: str  # 'huggingface', 'ollama', etc.
    model_id: str  # Unique identifier (e.g., "Qwen/Qwen2.5-32B")
    model_name: str  # Display name
    author: str  # Author/organization

    # Size information
    total_size_bytes: int
    file_count: int

    # Model specifications
    parameter_count: Optional[int] = None  # In billions (30 = 30B)
    architecture: Optional[str] = None  # "llama", "qwen", etc.
    format: List[str] = field(default_factory=list)
    quantization: Optional[str] = None  # "Q4_K_M", "fp16", etc.

    # Metadata
    tags: List[str] = field(default_factory=list)
    license: Optional[str] = None
    downloads: int = 0
    likes: int = 0
    created_at: str = ""
    updated_at: str = ""

    # URLs
    web_url: str = ""
    repo_url: str = ""

    # Additional source-specific data
    extra: Dict[str, Any] = field(default_factory=dict)
