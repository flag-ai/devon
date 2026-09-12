"""Storage status, clean, and export endpoints."""

from collections import Counter
from datetime import datetime, timedelta

from fastapi import APIRouter, Depends

from devon.api.dependencies import get_storage, verify_api_key
from devon.api.schemas import (
    CleanRequest,
    CleanResponse,
    ExportRequest,
    ExportResponse,
    StorageStatusResponse,
)
from devon.storage.organizer import ModelStorage

router = APIRouter(prefix="/api/v1", dependencies=[Depends(verify_api_key)])


@router.get("/status", response_model=StorageStatusResponse)
async def storage_status(storage: ModelStorage = Depends(get_storage)):
    """Get storage stats."""
    models = storage.list_local_models()
    sources_counter = Counter(m["source"] for m in models)

    sources = {}
    for source, count in sources_counter.items():
        source_size = sum(m["size_bytes"] for m in models if m["source"] == source)
        sources[source] = {"count": count, "size_bytes": source_size}

    return StorageStatusResponse(
        model_count=len(models),
        total_size_bytes=storage.get_total_size(),
        storage_path=str(storage.base_path),
        sources=sources,
    )


@router.post("/clean", response_model=CleanResponse)
async def clean_models(body: CleanRequest, storage: ModelStorage = Depends(get_storage)):
    """Clean unused or all models."""
    models = storage.list_local_models()
    to_remove = []

    if body.all:
        to_remove = models
    elif body.unused:
        cutoff = datetime.now() - timedelta(days=body.days)
        for model in models:
            last_used = model.get("last_used")
            if last_used is None:
                try:
                    downloaded = datetime.fromisoformat(model["downloaded_at"])
                except (ValueError, TypeError):
                    continue
                if downloaded < cutoff:
                    to_remove.append(model)
            else:
                try:
                    used_at = datetime.fromisoformat(last_used)
                except (ValueError, TypeError):
                    continue
                if used_at < cutoff:
                    to_remove.append(model)

    removed_ids = []
    freed = 0

    if not body.dry_run:
        for model in to_remove:
            if storage.delete_model(model["source"], model["model_id"]):
                removed_ids.append(f"{model['source']}/{model['model_id']}")
                freed += model["size_bytes"]
    else:
        for model in to_remove:
            removed_ids.append(f"{model['source']}/{model['model_id']}")
            freed += model["size_bytes"]

    return CleanResponse(
        removed=len(removed_ids),
        freed_bytes=freed,
        dry_run=body.dry_run,
        models=removed_ids,
    )


@router.post("/export", response_model=ExportResponse)
async def export_models(body: ExportRequest, storage: ModelStorage = Depends(get_storage)):
    """Export model list."""
    models = storage.list_local_models()

    if body.format == "kitt":
        content = [model["path"] for model in models]
    else:
        content = [
            {
                "source": model["source"],
                "model_id": model["model_id"],
                "path": model["path"],
                "size_bytes": model["size_bytes"],
                "downloaded_at": model["downloaded_at"],
                "files": model["files"],
            }
            for model in models
        ]

    return ExportResponse(format=body.format, count=len(models), content=content)
