from typing import Dict, Type, List

from devon.sources.base import ModelSource


class SourceRegistry:
    """Registry for discovering and managing model sources."""

    _sources: Dict[str, Type[ModelSource]] = {}

    @classmethod
    def register(cls, source_class: Type[ModelSource]) -> None:
        """Register a source class."""
        cls._sources[source_class.name()] = source_class

    @classmethod
    def get_source(cls, name: str) -> Type[ModelSource]:
        """Get source class by name."""
        if name not in cls._sources:
            available = ", ".join(cls._sources.keys())
            raise ValueError(f"Source '{name}' not found. Available sources: {available}")
        return cls._sources[name]

    @classmethod
    def list_available(cls) -> List[str]:
        """List all available (accessible) sources."""
        return [name for name, source_class in cls._sources.items() if source_class.is_available()]

    @classmethod
    def list_all(cls) -> List[str]:
        """List all registered sources."""
        return list(cls._sources.keys())


def register_source(source_class: Type[ModelSource]):
    """Decorator to auto-register source classes."""
    SourceRegistry.register(source_class)
    return source_class
