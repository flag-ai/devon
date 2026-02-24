"""Pydantic v2 request/response models for the DEVON API."""

from typing import Any, Dict, List, Optional

from pydantic import BaseModel, Field


# --- Health ---


class HealthResponse(BaseModel):
    status: str = "ok"
    version: str = "1.0.0"


# --- Model result (shared) ---


class ModelResult(BaseModel):
    model_config = {"from_attributes": True}

    source: str
    model_id: str
    model_name: str
    author: str
    total_size_bytes: int
    file_count: int
    parameter_count: Optional[int] = None
    architecture: Optional[str] = None
    format: List[str] = Field(default_factory=list)
    quantization: Optional[str] = None
    tags: List[str] = Field(default_factory=list)
    license: Optional[str] = None
    downloads: int = 0
    likes: int = 0
    created_at: str = ""
    updated_at: str = ""
    web_url: str = ""
    repo_url: str = ""


def metadata_to_result(metadata) -> ModelResult:
    """Convert a ModelMetadata dataclass to a ModelResult schema."""
    return ModelResult.model_validate(metadata)


# --- Search ---


class SearchResponse(BaseModel):
    query: Optional[str] = None
    source: str
    count: int
    results: List[ModelResult]


# --- Local models ---


class LocalModel(BaseModel):
    source: str
    model_id: str
    path: str
    size_bytes: int
    downloaded_at: str
    last_used: Optional[str] = None
    files: List[str] = Field(default_factory=list)
    metadata: Dict[str, Any] = Field(default_factory=dict)


class LocalModelsResponse(BaseModel):
    count: int
    models: List[LocalModel]


class ModelInfoResponse(BaseModel):
    local: Optional[LocalModel] = None
    remote: Optional[ModelResult] = None


class DeleteResponse(BaseModel):
    deleted: bool
    model_id: str
    source: str


# --- Download ---


class DownloadRequest(BaseModel):
    model_id: str
    source: str = "huggingface"
    force: bool = False
    include_patterns: Optional[List[str]] = None


class DownloadResponse(BaseModel):
    model_id: str
    source: str
    path: str
    files: List[str]
    size_bytes: int


class DownloadJobResponse(BaseModel):
    id: str
    model_id: str
    source: str
    status: str
    started_at: str
    completed_at: Optional[str] = None
    error: Optional[str] = None
    result: Optional[DownloadResponse] = None


class DownloadJobListResponse(BaseModel):
    count: int
    jobs: List[DownloadJobResponse]


class DownloadStartResponse(BaseModel):
    job: Optional[DownloadJobResponse] = None
    cached: Optional[DownloadResponse] = None


# --- Storage ---


class StorageStatusResponse(BaseModel):
    model_count: int
    total_size_bytes: int
    storage_path: str
    sources: Dict[str, Dict[str, Any]] = Field(default_factory=dict)


class CleanRequest(BaseModel):
    unused: bool = False
    days: int = 30
    all: bool = False
    dry_run: bool = False


class CleanResponse(BaseModel):
    removed: int
    freed_bytes: int
    dry_run: bool
    models: List[str]


class ExportRequest(BaseModel):
    format: str = "kitt"


class ExportResponse(BaseModel):
    format: str
    count: int
    content: Any


# --- Config ---


class ConfigResponse(BaseModel):
    config: Dict[str, Any]


class ConfigUpdateRequest(BaseModel):
    config: Dict[str, Any]


class SetupStatusResponse(BaseModel):
    configured: bool
    missing: List[str] = Field(default_factory=list)


class SecretsUpdateRequest(BaseModel):
    hf_token: Optional[str] = None
    api_key: Optional[str] = None


# --- Setup (first-run key provisioning) ---


class SetupCheckResponse(BaseModel):
    needs_setup: bool


class SetupKeyResponse(BaseModel):
    api_key: str


# --- Scan ---


class ScanRequest(BaseModel):
    reconcile: bool = False
    dry_run: bool = False
    path: Optional[str] = None


class ScanResultEntry(BaseModel):
    model_id: str
    source: str
    size_bytes: int
    status: str  # "new", "existing", "stale", "removed"


class ScanResponse(BaseModel):
    added: int
    existing: int
    stale: int
    removed: int
    models: List[ScanResultEntry]


# --- Errors ---


class ErrorResponse(BaseModel):
    detail: str
