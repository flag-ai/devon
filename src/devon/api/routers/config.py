"""Configuration management endpoints."""

import os

from fastapi import APIRouter, Depends, HTTPException

from devon.api.dependencies import get_settings, verify_api_key
from devon.api.schemas import (
    ConfigResponse,
    ConfigUpdateRequest,
    SecretsUpdateRequest,
    SetupStatusResponse,
)
from devon.config.settings import Settings

router = APIRouter(prefix="/api/v1", dependencies=[Depends(verify_api_key)])


@router.get("/config", response_model=ConfigResponse)
async def get_config(settings: Settings = Depends(get_settings)):
    """Return current configuration with secrets masked."""
    return ConfigResponse(config=settings.to_safe_dict())


@router.put("/config", response_model=ConfigResponse)
async def update_config(
    body: ConfigUpdateRequest,
    settings: Settings = Depends(get_settings),
):
    """Update configuration. Secrets in the payload are rejected — use PUT /config/secrets."""
    if "secrets" in body.config:
        raise HTTPException(
            status_code=400,
            detail="Secrets cannot be set via this endpoint. Use PUT /api/v1/config/secrets instead.",
        )

    settings.update(body.config)
    return ConfigResponse(config=settings.to_safe_dict())


@router.get("/config/setup-status", response_model=SetupStatusResponse)
async def setup_status(settings: Settings = Depends(get_settings)):
    """Check whether Devon has been configured.

    Returns configured=False with a list of recommended settings when
    no user config file exists on disk.
    """
    missing = []
    if not settings.is_configured:
        missing.append("storage.base_path")
        missing.append("secrets.hf_token")
    else:
        if not settings.get("secrets.hf_token") and not os.environ.get("HF_TOKEN"):
            missing.append("secrets.hf_token")

    return SetupStatusResponse(configured=settings.is_configured, missing=missing)


@router.put("/config/secrets")
async def update_secrets(
    body: SecretsUpdateRequest,
    settings: Settings = Depends(get_settings),
):
    """Write-only endpoint for secret values. Values are never returned."""
    secrets_update = {}
    if body.hf_token is not None:
        secrets_update["hf_token"] = body.hf_token
    if body.api_key is not None:
        secrets_update["api_key"] = body.api_key

    if secrets_update:
        settings.update({"secrets": secrets_update})

    return {"updated": list(secrets_update.keys())}
