"""Tests for the first-run API key setup flow."""

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


# ---------- GET /api/v1/setup/status ----------


class TestSetupStatus:
    @pytest.mark.anyio
    async def test_needs_setup_when_no_key_anywhere(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_API_KEY", raising=False)
        app = _make_app(monkeypatch, tmp_path)
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get("/api/v1/setup/status")
        assert resp.status_code == 200
        assert resp.json()["needs_setup"] is True

    @pytest.mark.anyio
    async def test_no_setup_when_env_key_set(self, tmp_path, monkeypatch):
        app = _make_app(monkeypatch, tmp_path, DEVON_API_KEY="some-secret")
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get("/api/v1/setup/status")
        assert resp.status_code == 200
        assert resp.json()["needs_setup"] is False

    @pytest.mark.anyio
    async def test_no_setup_when_config_key_set(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_API_KEY", raising=False)
        app = _make_app(monkeypatch, tmp_path)
        # Write a key into the config file
        app.state.settings.update({"secrets": {"api_key": "config-key"}})
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get("/api/v1/setup/status")
        assert resp.status_code == 200
        assert resp.json()["needs_setup"] is False


# ---------- POST /api/v1/setup ----------


class TestRunSetup:
    @pytest.mark.anyio
    async def test_generates_key_on_first_run(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_API_KEY", raising=False)
        app = _make_app(monkeypatch, tmp_path)
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.post("/api/v1/setup")
        assert resp.status_code == 201
        data = resp.json()
        assert "api_key" in data
        assert len(data["api_key"]) > 20  # token_urlsafe(32) is ~43 chars

    @pytest.mark.anyio
    async def test_409_on_second_run(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_API_KEY", raising=False)
        app = _make_app(monkeypatch, tmp_path)
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            first = await c.post("/api/v1/setup")
            assert first.status_code == 201
            second = await c.post("/api/v1/setup")
        assert second.status_code == 409
        assert "already configured" in second.json()["detail"]

    @pytest.mark.anyio
    async def test_409_when_env_key_exists(self, tmp_path, monkeypatch):
        app = _make_app(monkeypatch, tmp_path, DEVON_API_KEY="existing-key")
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.post("/api/v1/setup")
        assert resp.status_code == 409

    @pytest.mark.anyio
    async def test_generated_key_persists_in_config(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_API_KEY", raising=False)
        app = _make_app(monkeypatch, tmp_path)
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.post("/api/v1/setup")
        key = resp.json()["api_key"]
        # Verify persisted in settings
        assert app.state.settings.get("secrets.api_key") == key

    @pytest.mark.anyio
    async def test_500_and_rollback_on_write_failure(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_API_KEY", raising=False)
        app = _make_app(monkeypatch, tmp_path)
        # Make config dir read-only to trigger PermissionError on save
        app.state.settings.config_path.parent.mkdir(parents=True, exist_ok=True)
        app.state.settings.config_path.parent.chmod(0o444)
        transport = ASGITransport(app=app)
        try:
            async with AsyncClient(transport=transport, base_url="http://test") as c:
                resp = await c.post("/api/v1/setup")
                assert resp.status_code == 500
                assert "permissions" in resp.json()["detail"]
                # Key should be rolled back — setup still needed
                status_resp = await c.get("/api/v1/setup/status")
                assert status_resp.json()["needs_setup"] is True
        finally:
            # Restore permissions for cleanup
            app.state.settings.config_path.parent.chmod(0o755)

    @pytest.mark.anyio
    async def test_generated_key_authenticates(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_API_KEY", raising=False)
        app = _make_app(monkeypatch, tmp_path)
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.post("/api/v1/setup")
            key = resp.json()["api_key"]
            # Use the generated key to access an auth-protected endpoint
            config_resp = await c.get(
                "/api/v1/config",
                headers={"Authorization": f"Bearer {key}"},
            )
        assert config_resp.status_code == 200


# ---------- Three-tier auth behavior ----------


class TestThreeTierAuth:
    @pytest.mark.anyio
    async def test_config_key_authenticates_without_env(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_API_KEY", raising=False)
        app = _make_app(monkeypatch, tmp_path)
        app.state.settings.update({"secrets": {"api_key": "config-secret"}})
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get(
                "/api/v1/config",
                headers={"Authorization": "Bearer config-secret"},
            )
        assert resp.status_code == 200

    @pytest.mark.anyio
    async def test_env_key_takes_precedence_over_config(self, tmp_path, monkeypatch):
        app = _make_app(monkeypatch, tmp_path, DEVON_API_KEY="env-secret")
        app.state.settings.update({"secrets": {"api_key": "config-secret"}})
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            # Config key should NOT work when env key is set
            resp_config = await c.get(
                "/api/v1/config",
                headers={"Authorization": "Bearer config-secret"},
            )
            # Env key should work
            resp_env = await c.get(
                "/api/v1/config",
                headers={"Authorization": "Bearer env-secret"},
            )
        assert resp_config.status_code == 401
        assert resp_env.status_code == 200

    @pytest.mark.anyio
    async def test_503_setup_required_when_no_key(self, tmp_path, monkeypatch):
        monkeypatch.delenv("DEVON_API_KEY", raising=False)
        app = _make_app(monkeypatch, tmp_path)
        transport = ASGITransport(app=app)
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            resp = await c.get("/api/v1/config")
        assert resp.status_code == 503
        assert resp.json()["detail"] == "DEVON_SETUP_REQUIRED"
