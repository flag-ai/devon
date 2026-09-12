"""Search remote model sources."""

from typing import Optional

from fastapi import APIRouter, Depends

from devon.api.dependencies import get_source, verify_api_key
from devon.api.schemas import SearchResponse, metadata_to_result

router = APIRouter(prefix="/api/v1", dependencies=[Depends(verify_api_key)])


@router.get("/search", response_model=SearchResponse)
async def search(
    query: Optional[str] = None,
    source: str = "huggingface",
    provider: Optional[str] = None,
    params: Optional[str] = None,
    size: Optional[str] = None,
    format: Optional[str] = None,
    task: Optional[str] = None,
    license: Optional[str] = None,
    limit: int = 20,
):
    """Search remote sources. Query params mirror CLI filters."""
    filters = {}
    if provider:
        filters["author"] = provider
    if params:
        filters["params"] = params
    if size:
        filters["size"] = size
    if format:
        filters["format"] = format
    if task:
        filters["task"] = task
    if license:
        filters["license"] = license

    source_impl = get_source(source)
    results = source_impl.search(query=query, filters=filters, limit=limit)

    return SearchResponse(
        query=query,
        source=source,
        count=len(results),
        results=[metadata_to_result(m) for m in results],
    )
