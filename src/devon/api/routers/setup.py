"""First-run API key setup endpoints (unauthenticated)."""

import os
import secrets

from fastapi import APIRouter, Depends, HTTPException, status

from devon.api.dependencies import get_settings
from devon.api.schemas import SetupCheckResponse, SetupKeyResponse
from devon.config.settings import Settings

router = APIRouter(prefix="/api/v1/setup", tags=["setup"])


def _has_key(settings: Settings) -> bool:
    """Check if an API key is configured anywhere."""
    env_key = os.environ.get("DEVON_API_KEY", "")
    config_key = settings.get("secrets.api_key") or ""
    return bool(env_key) or bool(config_key)


@router.get("/status", response_model=SetupCheckResponse)
async def setup_status(settings: Settings = Depends(get_settings)):
    """Public probe — returns whether first-run setup is needed."""
    return SetupCheckResponse(needs_setup=not _has_key(settings))


@router.post("", response_model=SetupKeyResponse, status_code=status.HTTP_201_CREATED)
async def run_setup(settings: Settings = Depends(get_settings)):
    """Generate a new API key on first run.

    Returns the key exactly once (201). Returns 409 if a key already exists.
    """
    if _has_key(settings):
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="API key already configured",
        )

    api_key = secrets.token_urlsafe(32)

    try:
        settings.update({"secrets": {"api_key": api_key}})
    except OSError:
        # Roll back in-memory state so the setup can be retried.
        # Access _config directly to avoid another failed save().
        secrets_dict = settings._config.get("secrets", {})
        secrets_dict["api_key"] = None
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to persist API key — check config directory permissions",
        )

    return SetupKeyResponse(api_key=api_key)
