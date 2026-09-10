from dataclasses import asdict
from typing import Callable, List, Optional

from devon.sources.base import ModelSource
from devon.storage.organizer import ModelStorage


class DownloadManager:
    """Orchestrate model downloads."""

    def __init__(self, storage: Optional[ModelStorage] = None):
        self.storage = storage or ModelStorage()

    def download(
        self,
        source: ModelSource,
        model_id: str,
        force: bool = False,
        progress_callback: Optional[Callable] = None,
    ) -> List[str]:
        """
        Download a model and register it in storage.

        Args:
            source: The model source to download from
            model_id: Model identifier
            force: Whether to re-download if already present
            progress_callback: Optional progress callback

        Returns:
            List of downloaded file paths
        """
        source_name = source.name()

        if not force and self.storage.is_downloaded(source_name, model_id):
            entry = self.storage.get_model_entry(source_name, model_id)
            return entry["files"] if entry else []

        dest = self.storage.get_model_path(source_name, model_id)
        files = source.download_model(model_id, str(dest), progress_callback)

        # Get metadata for registration
        model_info = source.get_model_info(model_id)

        self.storage.register_model(
            source=source_name,
            model_id=model_id,
            metadata=asdict(model_info),
            files=files,
        )

        return files
