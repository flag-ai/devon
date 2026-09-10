"""Scan model directories and infer metadata for untracked models."""

import json
import logging
import re
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Set, Tuple

from devon.utils.format_utils import FORMAT_EXTENSIONS, detect_formats_from_files

logger = logging.getLogger(__name__)

# File extensions that indicate a directory contains model weights
MODEL_EXTENSIONS: Set[str] = set(FORMAT_EXTENSIONS.keys())

# First-level directories that map to registered source names
KNOWN_SOURCES: Set[str] = {"huggingface"}

# Architecture keywords detected from config.json model_type
ARCH_KEYWORDS = ["llama", "qwen", "mistral", "gpt", "bert", "gemma", "phi", "falcon", "mamba"]

# Quantization patterns detected from filenames and directory names
QUANT_PATTERNS = [
    "Q4_K_M",
    "Q5_K_M",
    "Q5_K_S",
    "Q8_0",
    "Q6_K",
    "Q4_K_S",
    "Q3_K_M",
    "Q3_K_S",
    "Q2_K",
    "IQ4_XS",
    "fp16",
    "bf16",
    "int8",
    "int4",
]

_PARAM_RE = re.compile(r"(\d+)[bB]\b")
_QUANT_RE = re.compile(
    r"(?<![a-zA-Z])(" + "|".join(re.escape(p) for p in QUANT_PATTERNS) + r")(?![a-zA-Z])"
)


class ModelScanner:
    """Scan a model directory tree and infer metadata for discovered models."""

    def scan(
        self,
        base_path: Path,
        existing_keys: Set[str],
    ) -> List[Dict]:
        """Find model directories not already in the manifest.

        Returns a list of entry dicts ready for register_model().
        """
        results = []
        for model_dir in self._find_model_dirs(base_path):
            source, model_id = self._infer_identity(base_path, model_dir)
            key = f"{source}::{model_id}"
            if key in existing_keys:
                continue

            metadata, files = self._infer_metadata(model_dir, source, model_id)
            entry = {
                "source": source,
                "model_id": model_id,
                "path": str(model_dir),
                "metadata": metadata,
                "files": files,
                "downloaded_at": datetime.now().isoformat(),
                "last_used": None,
                "size_bytes": metadata.get("total_size_bytes", 0),
            }
            results.append(entry)

        return results

    def find_stale(
        self,
        manifest_entries: Dict[str, Dict],
    ) -> List[str]:
        """Find manifest keys whose model paths no longer exist on disk."""
        stale = []
        for key, entry in manifest_entries.items():
            path = Path(entry["path"])
            if not path.exists():
                stale.append(key)
        return stale

    def _find_model_dirs(self, base_path: Path) -> List[Path]:
        """Walk the directory tree and return directories containing model files.

        Uses a deepest-directory-wins heuristic: if a directory contains model
        files but also has child directories with model files, only the children
        are returned. This avoids duplicate results when a model repo contains
        nested quantization variants.
        """
        model_dirs: List[Path] = []
        if not base_path.is_dir():
            return model_dirs

        for dirpath in base_path.rglob("*"):
            if not dirpath.is_dir():
                continue
            # Skip the base_path itself
            if dirpath == base_path:
                continue
            if self._is_model_dir(dirpath):
                # Don't include dirs that are parents of other model dirs
                # (we want the deepest model directory)
                if not any(
                    self._is_model_dir(child) for child in dirpath.iterdir() if child.is_dir()
                ):
                    model_dirs.append(dirpath)

        return model_dirs

    def _is_model_dir(self, path: Path) -> bool:
        """Check if a directory contains model weight files or a HF config."""
        for item in path.iterdir():
            if not item.is_file():
                continue
            if item.suffix in MODEL_EXTENSIONS:
                return True
            if item.name == "config.json":
                try:
                    data = json.loads(item.read_text())
                    if "model_type" in data:
                        return True
                except (json.JSONDecodeError, OSError):
                    pass
        return False

    def _infer_identity(
        self,
        base_path: Path,
        model_dir: Path,
    ) -> Tuple[str, str]:
        """Derive (source, model_id) from the directory path relative to base_path."""
        rel = model_dir.relative_to(base_path)
        parts = rel.parts

        if parts and parts[0] in KNOWN_SOURCES:
            source = parts[0]
            model_id = str(Path(*parts[1:])) if len(parts) > 1 else parts[0]
        else:
            source = "local"
            model_id = str(rel)

        return source, model_id

    def _infer_metadata(
        self,
        model_dir: Path,
        source: str,
        model_id: str,
    ) -> Tuple[Dict, List[str]]:
        """Infer as much metadata as possible from the model directory.

        Returns (metadata_dict, file_list).
        """
        # Collect files
        files: List[str] = []
        total_size = 0
        for item in model_dir.rglob("*"):
            if item.is_file():
                rel = str(item.relative_to(model_dir))
                files.append(rel)
                try:
                    total_size += item.stat().st_size
                except OSError:
                    pass

        # Format detection
        formats = detect_formats_from_files(files)

        # Architecture from config.json
        architecture = self._detect_architecture(model_dir)

        # Quantization from filenames and directory name
        quantization = self._detect_quantization(files, model_id)

        # Parameter count from directory name or config.json
        parameter_count = self._detect_parameter_count(model_dir, model_id)

        # Author and name from model_id
        id_parts = model_id.split("/")
        if len(id_parts) >= 2:
            author = id_parts[0]
            model_name = id_parts[-1]
        else:
            author = ""
            model_name = id_parts[-1]

        metadata = {
            "source": source,
            "model_id": model_id,
            "model_name": model_name,
            "author": author,
            "total_size_bytes": total_size,
            "file_count": len(files),
            "parameter_count": parameter_count,
            "architecture": architecture,
            "format": formats,
            "quantization": quantization,
            "tags": [],
            "license": None,
            "downloads": 0,
            "likes": 0,
            "created_at": "",
            "updated_at": "",
            "web_url": "",
            "repo_url": "",
        }

        return metadata, files

    def _detect_architecture(self, model_dir: Path) -> Optional[str]:
        """Read config.json model_type to determine architecture."""
        config_path = model_dir / "config.json"
        if not config_path.is_file():
            return None
        try:
            data = json.loads(config_path.read_text())
            model_type = data.get("model_type", "")
            for arch in ARCH_KEYWORDS:
                if arch in model_type.lower():
                    return arch
            # Return raw model_type if it doesn't match known keywords
            if model_type:
                return model_type.lower()
        except (json.JSONDecodeError, OSError):
            pass
        return None

    def _detect_quantization(self, files: List[str], model_id: str) -> Optional[str]:
        """Detect quantization from filenames or model ID using word boundaries."""
        text = " ".join(files) + " " + model_id
        match = _QUANT_RE.search(text)
        if match:
            return match.group(1)
        return None

    def _detect_parameter_count(self, model_dir: Path, model_id: str) -> Optional[int]:
        """Detect parameter count from directory name or config.json."""
        # Try from model_id / directory name
        match = _PARAM_RE.search(model_id)
        if match:
            return int(match.group(1))

        # Try from config.json (some models store this)
        config_path = model_dir / "config.json"
        if config_path.is_file():
            try:
                data = json.loads(config_path.read_text())
                # Some HF configs have num_parameters
                num_params = data.get("num_parameters")
                if num_params and isinstance(num_params, (int, float)):
                    return round(num_params / 1e9)
            except (json.JSONDecodeError, OSError):
                pass

        return None
