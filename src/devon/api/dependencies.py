"""Shared FastAPI dependencies for the DEVON API."""

import hmac
import os
from typing import Annotated

from fastapi import Depends, HTTPException, Request, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer

from devon.config.settings import Settings
from devon.sources.registry import SourceRegistry
from devon.storage.organizer import ModelStorage

# Optional bearer token — only enforced when DEVON_API_KEY is set
_bearer_scheme = HTTPBearer(auto_error=False)


def get_settings(request: Request) -> Settings:
    """Retrieve the Settings instance stored at startup."""
    return request.app.state.settings


def get_storage(request: Request) -> ModelStorage:
    """Retrieve the ModelStorage instance stored at startup."""
    return request.app.state.storage


def get_source(source_name: str = "huggingface"):
    """Instantiate a source plugin by name."""
    try:
        return SourceRegistry.get_source(source_name)()
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc))


async def verify_api_key(
    request: Request,
    credentials: Annotated[HTTPAuthorizationCredentials | None, Depends(_bearer_scheme)] = None,
) -> None:
    """Verify bearer token using three-tier auth.

    1. DEVON_API_KEY env var set → use it ("disable" allows unauthenticated).
    2. Env empty → check settings config file (secrets.api_key).
    3. Neither → raise 503 DEVON_SETUP_REQUIRED (triggers first-run flow).
    """
    expected = os.environ.get("DEVON_API_KEY", "")

    if not expected:
        # Tier 2: check config file
        settings: Settings = request.app.state.settings
        expected = settings.get("secrets.api_key") or ""

    if not expected:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="DEVON_SETUP_REQUIRED",
        )

    if expected == "disable":
        return

    if credentials is None or not hmac.compare_digest(credentials.credentials, expected):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid or missing API key",
            headers={"WWW-Authenticate": "Bearer"},
        )
