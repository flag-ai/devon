"""Tests for API authentication, security headers, CORS, and lifespan."""

import pytest
from httpx import ASGITransport, AsyncClient

from devon.config.settings import Settings
from devon.storage.organizer import ModelStorage


# ---------- helpers ----------


def _make_app(monkeypatch, tmp_path, **env_overrides):
    """Build a fresh app with temporary paths and optional env overrides."""
    config_path = tmp_path / "config.yaml"
    storage_path = tmp_path / "models"
    storage_path.mkdir(exist_ok=True)

    for key, value in env_overrides.items():
        if value is None:
            monkeypatch.delenv(key, raising=False)
        else:
            monkeypatch.setenv(key, value)

    from devon.api.app import create_app

    application = create_app()
    application.state.settings = Settings(config_path=config_path)
    application.state.storage = ModelStorage(base_path=storage_path)
    return application


# ---------- verify_api_key ----------


class TestVerifyApiKey:
    @pytest.mark.anyio
    async def test_503_when_api_key_not_set(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_API_KEY", raising=False)
        app = _make_app(monkeypatch, tmp_path)
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get("/api/v1/config")
        assert resp.status_code == 503
        assert resp.json()["detail"] == "DEVON_SETUP_REQUIRED"

    @pytest.mark.anyio
    async def test_passthrough_when_disabled(self, tmp_path, monkeypatch):
        app = _make_app(monkeypatch, tmp_path, DEVON_API_KEY="disable")
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get("/api/v1/config")
        assert resp.status_code == 200

    @pytest.mark.anyio
    async def test_401_when_no_credentials(self, tmp_path, monkeypatch):
        app = _make_app(monkeypatch, tmp_path, DEVON_API_KEY="real-secret")
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get("/api/v1/config")
        assert resp.status_code == 401
        assert "Invalid or missing API key" in resp.json()["detail"]

    @pytest.mark.anyio
    async def test_401_when_wrong_credentials(self, tmp_path, monkeypatch):
        app = _make_app(monkeypatch, tmp_path, DEVON_API_KEY="real-secret")
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get(
                "/api/v1/config",
                headers={"Authorization": "Bearer wrong-key"},
            )
        assert resp.status_code == 401

    @pytest.mark.anyio
    async def test_200_when_correct_credentials(self, tmp_path, monkeypatch):
        app = _make_app(monkeypatch, tmp_path, DEVON_API_KEY="real-secret")
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get(
                "/api/v1/config",
                headers={"Authorization": "Bearer real-secret"},
            )
        assert resp.status_code == 200


# ---------- Security headers middleware ----------


class TestSecurityHeaders:
    @pytest.fixture
    def app(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_ENABLE_HSTS", raising=False)
        monkeypatch.delenv("DEVON_FRAME_ANCESTORS", raising=False)
        return _make_app(monkeypatch, tmp_path, DEVON_API_KEY="disable")

    @pytest.fixture
    async def client(self, app):
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            yield c

    @pytest.mark.anyio
    async def test_security_headers_present(self, client):
        resp = await client.get("/health")
        assert resp.headers["X-Content-Type-Options"] == "nosniff"
        assert resp.headers["X-Frame-Options"] == "DENY"
        assert resp.headers["X-XSS-Protection"] == "1; mode=block"
        assert resp.headers["Referrer-Policy"] == "strict-origin-when-cross-origin"

    @pytest.mark.anyio
    async def test_frame_ancestors_replaces_xfo(self, tmp_path, monkeypatch):
        app = _make_app(
            monkeypatch,
            tmp_path,
            DEVON_API_KEY="disable",
            DEVON_FRAME_ANCESTORS="https://kitt.example.com",
        )
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get("/health")
        assert "X-Frame-Options" not in resp.headers
        assert (
            resp.headers["Content-Security-Policy"]
            == "frame-ancestors 'self' https://kitt.example.com"
        )

    @pytest.mark.anyio
    async def test_frame_ancestors_rejects_csp_injection(self, tmp_path, monkeypatch):
        app = _make_app(
            monkeypatch,
            tmp_path,
            DEVON_API_KEY="disable",
            DEVON_FRAME_ANCESTORS="https://ok.com; script-src 'unsafe-inline'",
        )
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get("/health")
        # All tokens contain invalid chars (semicolons, quotes); falls back to DENY
        assert resp.headers["X-Frame-Options"] == "DENY"
        assert "Content-Security-Policy" not in resp.headers

    @pytest.mark.anyio
    async def test_frame_ancestors_rejects_wildcard(self, tmp_path, monkeypatch):
        app = _make_app(
            monkeypatch,
            tmp_path,
            DEVON_API_KEY="disable",
            DEVON_FRAME_ANCESTORS="*",
        )
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get("/health")
        # Wildcard rejected, falls back to X-Frame-Options: DENY
        assert resp.headers["X-Frame-Options"] == "DENY"
        assert "Content-Security-Policy" not in resp.headers

    @pytest.mark.anyio
    async def test_no_hsts_by_default(self, client):
        resp = await client.get("/health")
        assert "Strict-Transport-Security" not in resp.headers

    @pytest.mark.anyio
    async def test_hsts_when_enabled(self, tmp_path, monkeypatch):
        app = _make_app(monkeypatch, tmp_path, DEVON_API_KEY="disable", DEVON_ENABLE_HSTS="1")
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get("/health")
        assert "Strict-Transport-Security" in resp.headers
        assert "max-age=" in resp.headers["Strict-Transport-Security"]


# ---------- CORS middleware ----------


class TestCORS:
    @pytest.mark.anyio
    async def test_default_origins_allowed(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_CORS_ORIGINS", raising=False)
        app = _make_app(monkeypatch, tmp_path, DEVON_API_KEY="disable")
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.options(
                "/health",
                headers={
                    "Origin": "http://localhost:5173",
                    "Access-Control-Request-Method": "GET",
                },
            )
        assert resp.headers.get("access-control-allow-origin") == "http://localhost:5173"

    @pytest.mark.anyio
    async def test_custom_origins_from_env(self, tmp_path, monkeypatch):
        app = _make_app(
            monkeypatch,
            tmp_path,
            DEVON_API_KEY="disable",
            DEVON_CORS_ORIGINS="https://example.com",
        )
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.options(
                "/health",
                headers={
                    "Origin": "https://example.com",
                    "Access-Control-Request-Method": "GET",
                },
            )
        assert resp.headers.get("access-control-allow-origin") == "https://example.com"

    @pytest.mark.anyio
    async def test_unauthorized_origin_rejected(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_CORS_ORIGINS", raising=False)
        app = _make_app(monkeypatch, tmp_path, DEVON_API_KEY="disable")
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.options(
                "/health",
                headers={
                    "Origin": "https://evil.example.com",
                    "Access-Control-Request-Method": "GET",
                },
            )
        assert "access-control-allow-origin" not in resp.headers


# ---------- Lifespan ----------


class TestLifespan:
    @pytest.mark.anyio
    async def test_lifespan_initializes_state(self, tmp_path, monkeypatch):
        monkeypatch.setenv("DEVON_API_KEY", "disable")
        monkeypatch.delenv("DEVON_CONFIG_PATH", raising=False)
        monkeypatch.delenv("DEVON_STORAGE_PATH", raising=False)

        from devon.api.app import create_app, lifespan

        app = create_app()
        async with lifespan(app):
            assert hasattr(app.state, "settings")
            assert hasattr(app.state, "storage")
            assert isinstance(app.state.settings, Settings)
            assert isinstance(app.state.storage, ModelStorage)

    @pytest.mark.anyio
    async def test_lifespan_respects_config_path(self, tmp_path, monkeypatch):
        config_file = tmp_path / "custom.yaml"
        monkeypatch.setenv("DEVON_API_KEY", "disable")
        monkeypatch.setenv("DEVON_CONFIG_PATH", str(config_file))
        monkeypatch.delenv("DEVON_STORAGE_PATH", raising=False)

        from devon.api.app import create_app, lifespan

        app = create_app()
        async with lifespan(app):
            assert app.state.settings.config_path == config_file

    @pytest.mark.anyio
    async def test_lifespan_respects_storage_path(self, tmp_path, monkeypatch):
        storage_dir = tmp_path / "custom_models"
        storage_dir.mkdir()
        monkeypatch.setenv("DEVON_API_KEY", "disable")
        monkeypatch.setenv("DEVON_STORAGE_PATH", str(storage_dir))
        monkeypatch.delenv("DEVON_CONFIG_PATH", raising=False)

        from devon.api.app import create_app, lifespan

        app = create_app()
        async with lifespan(app):
            assert app.state.storage.base_path == storage_dir
