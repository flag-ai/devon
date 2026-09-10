"""Model listing, info, and deletion endpoints."""

import logging
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException

from devon.api.dependencies import get_source, get_storage, verify_api_key
from devon.api.schemas import (
    DeleteResponse,
    LocalModel,
    LocalModelsResponse,
    ModelInfoResponse,
    metadata_to_result,
)
from devon.storage.organizer import ModelStorage

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/v1", dependencies=[Depends(verify_api_key)])


@router.get("/models", response_model=LocalModelsResponse)
async def list_models(
    storage: ModelStorage = Depends(get_storage),
    source: Optional[str] = None,
):
    """List all locally downloaded models."""
    models = storage.list_local_models(source=source)
    return LocalModelsResponse(
        count=len(models),
        models=[LocalModel(**m) for m in models],
    )


@router.get("/models/{source}/{model_id:path}", response_model=ModelInfoResponse)
async def model_info(
    source: str,
    model_id: str,
    storage: ModelStorage = Depends(get_storage),
):
    """Get info for a model (local entry + remote metadata)."""
    local_entry = storage.get_model_entry(source, model_id)
    local = LocalModel(**local_entry) if local_entry else None

    remote = None
    try:
        source_impl = get_source(source)
        metadata = source_impl.get_model_info(model_id)
        remote = metadata_to_result(metadata)
    except HTTPException:
        raise
    except (ConnectionError, TimeoutError) as exc:
        logger.warning("Remote lookup failed for %s/%s: %s", source, model_id, exc)
    except Exception as exc:
        logger.warning("Unexpected error during remote lookup for %s/%s: %s", source, model_id, exc)

    if local is None and remote is None:
        raise HTTPException(status_code=404, detail=f"Model not found: {source}/{model_id}")

    return ModelInfoResponse(local=local, remote=remote)


@router.delete("/models/{source}/{model_id:path}", response_model=DeleteResponse)
async def delete_model(
    source: str,
    model_id: str,
    storage: ModelStorage = Depends(get_storage),
):
    """Remove a model from local storage."""
    deleted = storage.delete_model(source, model_id)
    if not deleted:
        raise HTTPException(status_code=404, detail=f"Model not found: {source}/{model_id}")
    return DeleteResponse(deleted=True, model_id=model_id, source=source)
