"""Scan endpoint to discover untracked models."""

from pathlib import Path
from typing import List

from fastapi import APIRouter, Depends

from devon.api.dependencies import get_storage, verify_api_key
from devon.api.schemas import ScanRequest, ScanResponse, ScanResultEntry
from devon.storage.organizer import ModelStorage
from devon.storage.scanner import ModelScanner

router = APIRouter(prefix="/api/v1", dependencies=[Depends(verify_api_key)])


@router.post("/scan", response_model=ScanResponse)
async def scan_models(
    body: ScanRequest,
    storage: ModelStorage = Depends(get_storage),
):
    """Scan model directory to discover untracked models.

    Walks the model directory tree, identifies model files, and registers
    any not already in the manifest. Use reconcile=true to also remove
    entries whose files no longer exist on disk.
    """
    scan_path = Path(body.path) if body.path else storage.base_path
    scanner = ModelScanner()

    existing_keys = set(storage.index.keys())
    new_entries = scanner.scan(scan_path, existing_keys)
    stale_keys = scanner.find_stale(storage.index)

    results: List[ScanResultEntry] = []
    added = 0
    removed = 0

    for entry in new_entries:
        results.append(ScanResultEntry(
            model_id=entry["model_id"],
            source=entry["source"],
            size_bytes=entry["size_bytes"],
            status="new",
        ))
        added += 1
        if not body.dry_run:
            storage.index[f"{entry['source']}::{entry['model_id']}"] = entry

    for key in stale_keys:
        entry = storage.index[key]
        if body.reconcile and not body.dry_run:
            status = "removed"
            removed += 1
        else:
            status = "stale"
        results.append(ScanResultEntry(
            model_id=entry["model_id"],
            source=entry["source"],
            size_bytes=entry["size_bytes"],
            status=status,
        ))

    if body.reconcile and not body.dry_run:
        for key in stale_keys:
            del storage.index[key]

    if not body.dry_run and (added > 0 or removed > 0):
        storage._save_index()

    return ScanResponse(
        added=added,
        existing=len(existing_keys),
        stale=len(stale_keys),
        removed=removed,
        models=results,
    )
