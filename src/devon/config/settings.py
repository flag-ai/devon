import copy
import os
from pathlib import Path
from typing import Any, Dict, List, Optional

import yaml


DEFAULT_CONFIG = {
    "storage": {
        "base_path": str(Path.home() / ".cache" / "devon" / "models"),
        "max_size_gb": None,
    },
    "download": {
        "resume": True,
        "verify_checksums": True,
    },
    "sources": {
        "default": "huggingface",
        "enabled": ["huggingface"],
    },
    "search": {
        "default_limit": 20,
        "sort_by": "downloads",
    },
    "display": {
        "color": True,
    },
    "secrets": {
        "hf_token": None,
        "api_key": None,
    },
}

SECRET_KEYS = {"secrets.hf_token", "secrets.api_key"}
SECRET_MASK = "\u2022\u2022\u2022\u2022\u2022\u2022"


class Settings:
    """User settings manager."""

    def __init__(self, config_path: Optional[Path] = None):
        if config_path is None:
            config_path = Path.home() / ".config" / "devon" / "config.yaml"

        self.config_path = Path(config_path)
        self._config = self._load()

    def _load(self) -> Dict[str, Any]:
        """Load config, merging with defaults."""
        config = dict(DEFAULT_CONFIG)

        if self.config_path.exists():
            with open(self.config_path) as f:
                user_config = yaml.safe_load(f) or {}
            config = self._deep_merge(config, user_config)

        return config

    def _deep_merge(self, base: Dict, override: Dict) -> Dict:
        """Deep merge two dicts."""
        result = dict(base)
        for key, value in override.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                result[key] = self._deep_merge(result[key], value)
            else:
                result[key] = value
        return result

    def get(self, key: str, default: Any = None) -> Any:
        """Get a config value by dot-separated key."""
        parts = key.split(".")
        value = self._config
        for part in parts:
            if isinstance(value, dict) and part in value:
                value = value[part]
            else:
                return default
        return value

    @property
    def storage_path(self) -> Path:
        """Get the model storage path."""
        path_str = self.get("storage.base_path")
        return Path(path_str).expanduser()

    @property
    def default_source(self) -> str:
        """Get the default source name."""
        return self.get("sources.default", "huggingface")

    @property
    def enabled_sources(self) -> List[str]:
        """Get list of enabled sources."""
        return self.get("sources.enabled", ["huggingface"])

    @property
    def search_limit(self) -> int:
        """Get default search result limit."""
        return self.get("search.default_limit", 20)

    @property
    def is_configured(self) -> bool:
        """True if a user config file exists on disk."""
        return self.config_path.exists()

    def update(self, updates: Dict[str, Any]) -> None:
        """Merge partial updates into config and persist to disk.

        Accepts a nested dict that is deep-merged over the current config.
        Secret keys are handled here — env vars for HF_TOKEN are also updated.
        """
        self._config = self._deep_merge(self._config, updates)

        # Sync HF_TOKEN env var when set via config
        hf_token = self.get("secrets.hf_token")
        if hf_token:
            os.environ["HF_TOKEN"] = hf_token

        self.save()

    def to_safe_dict(self) -> Dict[str, Any]:
        """Return the full config with secret values masked."""
        result = copy.deepcopy(self._config)
        secrets = result.get("secrets", {})
        for key in list(secrets.keys()):
            if secrets[key]:
                secrets[key] = SECRET_MASK
        return result

    def save(self) -> None:
        """Save current config to disk."""
        self.config_path.parent.mkdir(parents=True, exist_ok=True, mode=0o700)
        with open(self.config_path, "w") as f:
            yaml.dump(self._config, f, default_flow_style=False)
        self.config_path.chmod(0o600)
