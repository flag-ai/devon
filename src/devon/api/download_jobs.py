"""In-memory download job tracker for async model downloads."""

import asyncio
import logging
from dataclasses import asdict, dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional
from uuid import uuid4

from devon.sources.base import ModelSource
from devon.storage.organizer import ModelStorage

logger = logging.getLogger(__name__)


class JobStatus(str, Enum):
    downloading = "downloading"
    completed = "completed"
    failed = "failed"


@dataclass
class DownloadJob:
    id: str
    model_id: str
    source: str
    status: JobStatus
    started_at: str
    include_patterns: Optional[List[str]] = None
    force: bool = False
    completed_at: Optional[str] = None
    error: Optional[str] = None
    result: Optional[Dict[str, Any]] = field(default=None)


class DownloadJobManager:
    """Track in-flight and completed download jobs."""

    def __init__(self) -> None:
        self._jobs: Dict[str, DownloadJob] = {}

    def create_job(
        self,
        model_id: str,
        source: str,
        include_patterns: Optional[List[str]] = None,
        force: bool = False,
    ) -> DownloadJob:
        job = DownloadJob(
            id=uuid4().hex,
            model_id=model_id,
            source=source,
            status=JobStatus.downloading,
            started_at=datetime.now().isoformat(),
            include_patterns=include_patterns,
            force=force,
        )
        self._jobs[job.id] = job
        return job

    def get_job(self, job_id: str) -> Optional[DownloadJob]:
        return self._jobs.get(job_id)

    def list_jobs(self) -> List[DownloadJob]:
        return sorted(self._jobs.values(), key=lambda j: j.started_at, reverse=True)

    def has_active_job(self, source: str, model_id: str) -> Optional[DownloadJob]:
        """Return an active (downloading) job for the given source+model, or None."""
        for job in self._jobs.values():
            if job.source == source and job.model_id == model_id and job.status == JobStatus.downloading:
                return job
        return None

    async def run_download(
        self,
        job: DownloadJob,
        source_impl: ModelSource,
        storage: ModelStorage,
    ) -> None:
        """Execute download in a thread and update job status on completion."""
        try:
            dest = storage.get_model_path(job.source, job.model_id)
            allow_patterns = job.include_patterns if job.include_patterns else None

            files = await asyncio.to_thread(
                source_impl.download_model,
                job.model_id,
                str(dest),
                allow_patterns=allow_patterns,
            )

            model_info = await asyncio.to_thread(source_impl.get_model_info, job.model_id)
            metadata_dict = asdict(model_info)
            storage.register_model(
                source=job.source,
                model_id=job.model_id,
                metadata=metadata_dict,
                files=files,
            )

            entry = storage.get_model_entry(job.source, job.model_id)
            job.status = JobStatus.completed
            job.completed_at = datetime.now().isoformat()
            job.result = {
                "path": entry["path"] if entry else str(dest),
                "files": entry["files"] if entry else files,
                "size_bytes": entry["size_bytes"] if entry else 0,
            }
        except Exception as exc:
            logger.exception("Download failed for %s/%s", job.source, job.model_id)
            job.status = JobStatus.failed
            job.completed_at = datetime.now().isoformat()
            job.error = str(exc)
