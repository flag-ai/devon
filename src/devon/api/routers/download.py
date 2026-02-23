"""Async model download endpoints with job tracking."""

import asyncio

from fastapi import APIRouter, Depends, HTTPException, Request
from fastapi.responses import JSONResponse

from devon.api.dependencies import get_source, get_storage, verify_api_key
from devon.api.download_jobs import DownloadJobManager
from devon.api.schemas import (
    DownloadJobListResponse,
    DownloadJobResponse,
    DownloadRequest,
    DownloadResponse,
    DownloadStartResponse,
)
from devon.storage.organizer import ModelStorage

router = APIRouter(prefix="/api/v1", dependencies=[Depends(verify_api_key)])


def _job_to_response(job) -> DownloadJobResponse:
    result = None
    if job.result:
        result = DownloadResponse(
            model_id=job.model_id,
            source=job.source,
            path=job.result["path"],
            files=job.result["files"],
            size_bytes=job.result["size_bytes"],
        )
    return DownloadJobResponse(
        id=job.id,
        model_id=job.model_id,
        source=job.source,
        status=job.status.value,
        started_at=job.started_at,
        completed_at=job.completed_at,
        error=job.error,
        result=result,
    )


def _get_job_manager(request: Request) -> DownloadJobManager:
    return request.app.state.download_jobs


@router.post("/downloads")
async def start_download(
    body: DownloadRequest,
    request: Request,
    storage: ModelStorage = Depends(get_storage),
):
    """Start a model download.

    Returns 200 with cached result if already downloaded (and not forced).
    Returns 202 with job info for new or in-progress downloads.
    Validation failures (model not found) return 404 without creating a job.
    """
    jobs = _get_job_manager(request)
    source_impl = get_source(body.source)

    # Check for cached download (not forced)
    if not body.force and storage.is_downloaded(body.source, body.model_id):
        existing = storage.get_model_entry(body.source, body.model_id)
        if existing:
            cached = DownloadResponse(
                model_id=body.model_id,
                source=body.source,
                path=existing["path"],
                files=existing["files"],
                size_bytes=existing["size_bytes"],
            )
            return JSONResponse(
                status_code=200,
                content=DownloadStartResponse(cached=cached).model_dump(),
            )

    # Duplicate active download — return existing job
    active = jobs.has_active_job(body.source, body.model_id)
    if active:
        return JSONResponse(
            status_code=202,
            content=DownloadStartResponse(job=_job_to_response(active)).model_dump(),
        )

    # Validate model exists (fast HF API call) before creating a job
    try:
        source_impl.get_model_info(body.model_id)
    except Exception as exc:
        raise HTTPException(status_code=404, detail=f"Model not found: {exc}")

    # Create job and launch background download
    job = jobs.create_job(
        model_id=body.model_id,
        source=body.source,
        include_patterns=body.include_patterns,
        force=body.force,
    )
    asyncio.create_task(jobs.run_download(job, source_impl, storage))

    return JSONResponse(
        status_code=202,
        content=DownloadStartResponse(job=_job_to_response(job)).model_dump(),
    )


@router.get("/downloads", response_model=DownloadJobListResponse)
async def list_downloads(request: Request):
    """List all tracked download jobs, newest first."""
    jobs = _get_job_manager(request)
    all_jobs = jobs.list_jobs()
    return DownloadJobListResponse(
        count=len(all_jobs),
        jobs=[_job_to_response(j) for j in all_jobs],
    )


@router.get("/downloads/{job_id}", response_model=DownloadJobResponse)
async def get_download(job_id: str, request: Request):
    """Get status of a single download job."""
    jobs = _get_job_manager(request)
    job = jobs.get_job(job_id)
    if not job:
        raise HTTPException(status_code=404, detail=f"Job not found: {job_id}")
    return _job_to_response(job)


@router.post("/downloads/{job_id}/restart", response_model=DownloadStartResponse)
async def restart_download(
    job_id: str,
    request: Request,
    storage: ModelStorage = Depends(get_storage),
):
    """Restart a failed download. Creates a new job with force=True."""
    jobs = _get_job_manager(request)
    old_job = jobs.get_job(job_id)
    if not old_job:
        raise HTTPException(status_code=404, detail=f"Job not found: {job_id}")
    if old_job.status.value != "failed":
        raise HTTPException(status_code=400, detail="Only failed jobs can be restarted")

    source_impl = get_source(old_job.source)

    # Validate model still exists
    try:
        source_impl.get_model_info(old_job.model_id)
    except Exception as exc:
        raise HTTPException(status_code=404, detail=f"Model not found: {exc}")

    new_job = jobs.create_job(
        model_id=old_job.model_id,
        source=old_job.source,
        include_patterns=old_job.include_patterns,
        force=True,
    )
    asyncio.create_task(jobs.run_download(new_job, source_impl, storage))

    return DownloadStartResponse(job=_job_to_response(new_job))
