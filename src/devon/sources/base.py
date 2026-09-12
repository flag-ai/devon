from abc import ABC, abstractmethod
from typing import List, Dict, Any, Optional

from devon.models.model_info import ModelMetadata


class ModelSource(ABC):
    """Abstract base class for model sources."""

    @classmethod
    @abstractmethod
    def name(cls) -> str:
        """Source identifier (e.g., 'huggingface')."""
        pass

    @classmethod
    @abstractmethod
    def is_available(cls) -> bool:
        """Check if source is accessible."""
        pass

    @abstractmethod
    def search(
        self,
        query: Optional[str] = None,
        filters: Optional[Dict[str, Any]] = None,
        limit: int = 20,
    ) -> List[ModelMetadata]:
        """
        Search for models matching criteria.

        Args:
            query: Text search query
            filters: Dict of filters (author, params, size, format, etc.)
            limit: Max results to return

        Returns:
            List of ModelMetadata objects
        """
        pass

    @abstractmethod
    def get_model_info(self, model_id: str) -> ModelMetadata:
        """Get detailed info about specific model."""
        pass

    @abstractmethod
    def download_model(
        self,
        model_id: str,
        dest_path: str,
        progress_callback=None,
        allow_patterns: Optional[List[str]] = None,
    ) -> List[str]:
        """
        Download all files for a model.

        Args:
            model_id: Model identifier
            dest_path: Destination directory
            progress_callback: Optional callback for progress updates
            allow_patterns: Optional glob patterns to filter files (e.g., ["*Q4_K_M*"])

        Returns:
            List of downloaded file paths
        """
        pass
