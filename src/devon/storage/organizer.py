import json
import logging
import shutil
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional

logger = logging.getLogger(__name__)


class ModelStorage:
    """Manage local model storage."""

    def __init__(self, base_path: Optional[Path] = None):
        """
        Initialize storage manager.

        Reads base_path from ~/.config/devon/config.yaml if not provided.
        Falls back to ~/.cache/devon/models/ if no config exists.
        """
        if base_path is None:
            from devon.config.settings import Settings

            settings = Settings()
            base_path = settings.storage_path

        self.base_path = Path(base_path)
        self.index_file = self.base_path / "manifest.json"

        self.base_path.mkdir(parents=True, exist_ok=True, mode=0o700)
        self._migrate_index()
        self.index = self._load_index()

    def _validate_path(self, path: Path) -> Path:
        """Ensure path resolves within base_path. Raises ValueError if not."""
        resolved = path.resolve()
        base_resolved = self.base_path.resolve()
        if not resolved.is_relative_to(base_resolved):
            raise ValueError(f"Path traversal detected: {path} resolves outside {self.base_path}")
        return resolved

    def get_model_path(self, source: str, model_id: str) -> Path:
        """Get storage path for a model."""
        path = self.base_path / source / model_id
        self._validate_path(path)
        return path

    def register_model(
        self,
        source: str,
        model_id: str,
        metadata: Dict,
        files: List[str],
    ) -> None:
        """Register downloaded model."""
        path = self.get_model_path(source, model_id)
        size_bytes = 0
        for f in files:
            try:
                size_bytes += (path / f).stat().st_size
            except (FileNotFoundError, OSError):
                pass

        # Remove non-serializable data from metadata
        clean_metadata = {}
        for k, v in metadata.items():
            if k == "extra":
                continue
            try:
                json.dumps(v)
                clean_metadata[k] = v
            except (TypeError, ValueError):
                logger.debug("Metadata key %r not JSON-serializable, converting to str", k)
                clean_metadata[k] = str(v)

        entry = {
            "source": source,
            "model_id": model_id,
            "path": str(path),
            "metadata": clean_metadata,
            "files": files,
            "downloaded_at": datetime.now().isoformat(),
            "last_used": None,
            "size_bytes": size_bytes,
        }

        key = f"{source}::{model_id}"
        self.index[key] = entry
        self._save_index()

    def list_local_models(self, source: Optional[str] = None) -> List[Dict]:
        """List all locally downloaded models."""
        models = []
        for entry in self.index.values():
            if source is None or entry["source"] == source:
                models.append(entry)
        return models

    def is_downloaded(self, source: str, model_id: str) -> bool:
        """Check if model is downloaded."""
        return f"{source}::{model_id}" in self.index

    def get_model_entry(self, source: str, model_id: str) -> Optional[Dict]:
        """Get entry for a model."""
        return self.index.get(f"{source}::{model_id}")

    def delete_model(self, source: str, model_id: str) -> bool:
        """Delete model from disk and index."""
        key = f"{source}::{model_id}"
        if key not in self.index:
            return False

        path = Path(self.index[key]["path"])
        self._validate_path(path)
        if path.exists():
            shutil.rmtree(path)

        del self.index[key]
        self._save_index()
        return True

    def get_total_size(self) -> int:
        """Get total size of all models."""
        return sum(entry["size_bytes"] for entry in self.index.values())

    def mark_used(self, source: str, model_id: str) -> None:
        """Mark a model as recently used."""
        key = f"{source}::{model_id}"
        if key in self.index:
            self.index[key]["last_used"] = datetime.now().isoformat()
            self._save_index()

    def _migrate_index(self) -> None:
        """Migrate legacy index.json to manifest.json if needed."""
        legacy_index = self.base_path.parent / "index.json"
        if legacy_index.exists() and not self.index_file.exists():
            logger.info("Migrating index.json → manifest.json")
            data = json.loads(legacy_index.read_text())
            self.index_file.parent.mkdir(parents=True, exist_ok=True)
            with open(self.index_file, "w") as f:
                json.dump(data, f, indent=2)
            self.index_file.chmod(0o600)
            legacy_index.unlink()
            logger.info("Migration complete — deleted legacy index.json")

    def _load_index(self) -> Dict:
        """Load index from disk."""
        if self.index_file.exists():
            with open(self.index_file) as f:
                data = json.load(f)
            if not isinstance(data, dict):
                logger.warning("Manifest is not a JSON object, resetting to empty index")
                return {}
            return data
        return {}

    def _save_index(self) -> None:
        """Save index to disk."""
        self.index_file.parent.mkdir(parents=True, exist_ok=True)
        with open(self.index_file, "w") as f:
            json.dump(self.index, f, indent=2)
        self.index_file.chmod(0o600)
