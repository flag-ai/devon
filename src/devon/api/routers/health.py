"""Health check endpoint."""

from fastapi import APIRouter

from devon.api.schemas import HealthResponse

router = APIRouter()


@router.get("/health", response_model=HealthResponse)
async def health():
    """Health check — no auth required."""
    return HealthResponse()
