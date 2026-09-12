from typing import Any, Dict, List, Optional

from devon.models.model_info import ModelMetadata


class ModelFilter:
    """Filter models based on various criteria."""

    def __init__(self, filters: Optional[Dict[str, Any]] = None):
        self.filters = filters or {}

    def matches(self, model: ModelMetadata) -> bool:
        """Check if a model matches all filters."""
        for key, value in self.filters.items():
            checker = getattr(self, f"_check_{key}", None)
            if checker and not checker(model, value):
                return False
        return True

    def apply(self, models: List[ModelMetadata]) -> List[ModelMetadata]:
        """Filter a list of models."""
        return [m for m in models if self.matches(m)]

    def _check_author(self, model: ModelMetadata, value: str) -> bool:
        return model.author.lower() == value.lower()

    def _check_format(self, model: ModelMetadata, value: str | list[str]) -> bool:
        values = [value] if isinstance(value, str) else value
        return any(f in model.format for f in values)

    def _check_architecture(self, model: ModelMetadata, value: str) -> bool:
        return model.architecture is not None and model.architecture.lower() == value.lower()

    def _check_min_downloads(self, model: ModelMetadata, value: int) -> bool:
        return model.downloads >= value

    def _check_license(self, model: ModelMetadata, value: str) -> bool:
        return model.license is not None and model.license.lower() == value.lower()
