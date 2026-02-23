"""Tests for the configuration API endpoints."""

import pytest
from httpx import ASGITransport, AsyncClient

from devon.config.settings import SECRET_MASK, Settings
from devon.storage.organizer import ModelStorage


@pytest.fixture
def app(tmp_path, monkeypatch):
    """Create a test app with temp config and storage."""
    config_path = tmp_path / "config.yaml"
    storage_path = tmp_path / "models"
    storage_path.mkdir()

    monkeypatch.setenv("DEVON_API_KEY", "disable")
    monkeypatch.delenv("HF_TOKEN", raising=False)

    from devon.api.app import create_app

    application = create_app()
    application.state.settings = Settings(config_path=config_path)
    application.state.storage = ModelStorage(base_path=storage_path)

    return application


@pytest.fixture
async def client(app):
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as c:
        yield c


class TestGetConfig:
    @pytest.mark.anyio
    async def test_returns_config(self, client):
        resp = await client.get("/api/v1/config")
        assert resp.status_code == 200
        data = resp.json()
        assert "config" in data
        assert "storage" in data["config"]
        assert "download" in data["config"]
        assert "sources" in data["config"]
        assert "search" in data["config"]
        assert "display" in data["config"]
        assert "secrets" in data["config"]

    @pytest.mark.anyio
    async def test_masks_secrets(self, client):
        await client.put("/api/v1/config/secrets", json={"hf_token": "my-token"})

        resp = await client.get("/api/v1/config")
        data = resp.json()
        assert data["config"]["secrets"]["hf_token"] == SECRET_MASK

    @pytest.mark.anyio
    async def test_none_secrets_not_masked(self, client):
        resp = await client.get("/api/v1/config")
        data = resp.json()
        assert data["config"]["secrets"]["hf_token"] is None


class TestUpdateConfig:
    @pytest.mark.anyio
    async def test_update_storage_path(self, client):
        resp = await client.put(
            "/api/v1/config",
            json={"config": {"storage": {"base_path": "/new/path"}}},
        )
        assert resp.status_code == 200

        resp = await client.get("/api/v1/config")
        assert resp.json()["config"]["storage"]["base_path"] == "/new/path"

    @pytest.mark.anyio
    async def test_strips_secrets_from_config_update(self, client):
        resp = await client.put(
            "/api/v1/config",
            json={"config": {"secrets": {"hf_token": "should-be-ignored"}}},
        )
        assert resp.status_code == 200

        resp = await client.get("/api/v1/config")
        assert resp.json()["config"]["secrets"]["hf_token"] is None

    @pytest.mark.anyio
    async def test_deep_merge_preserves_siblings(self, client):
        resp = await client.put(
            "/api/v1/config",
            json={"config": {"storage": {"max_size_gb": 500}}},
        )
        assert resp.status_code == 200

        resp = await client.get("/api/v1/config")
        config = resp.json()["config"]
        assert config["storage"]["max_size_gb"] == 500
        assert config["storage"]["base_path"] is not None


class TestSetupStatus:
    @pytest.mark.anyio
    async def test_not_configured_initially(self, client):
        resp = await client.get("/api/v1/config/setup-status")
        assert resp.status_code == 200
        data = resp.json()
        assert data["configured"] is False
        assert "storage.base_path" in data["missing"]
        assert "secrets.hf_token" in data["missing"]

    @pytest.mark.anyio
    async def test_configured_after_save(self, client):
        await client.put(
            "/api/v1/config",
            json={"config": {"display": {"color": True}}},
        )

        resp = await client.get("/api/v1/config/setup-status")
        data = resp.json()
        assert data["configured"] is True

    @pytest.mark.anyio
    async def test_missing_hf_token_when_configured(self, client):
        await client.put(
            "/api/v1/config",
            json={"config": {"display": {"color": True}}},
        )

        resp = await client.get("/api/v1/config/setup-status")
        data = resp.json()
        assert data["configured"] is True
        assert "secrets.hf_token" in data["missing"]


class TestUpdateSecrets:
    @pytest.mark.anyio
    async def test_set_hf_token(self, client):
        resp = await client.put(
            "/api/v1/config/secrets",
            json={"hf_token": "test-token"},
        )
        assert resp.status_code == 200
        assert "hf_token" in resp.json()["updated"]

    @pytest.mark.anyio
    async def test_set_api_key(self, client):
        resp = await client.put(
            "/api/v1/config/secrets",
            json={"api_key": "test-key"},
        )
        assert resp.status_code == 200
        assert "api_key" in resp.json()["updated"]

    @pytest.mark.anyio
    async def test_set_both(self, client):
        resp = await client.put(
            "/api/v1/config/secrets",
            json={"hf_token": "tok", "api_key": "key"},
        )
        assert resp.status_code == 200
        updated = resp.json()["updated"]
        assert "hf_token" in updated
        assert "api_key" in updated

    @pytest.mark.anyio
    async def test_empty_update(self, client):
        resp = await client.put("/api/v1/config/secrets", json={})
        assert resp.status_code == 200
        assert resp.json()["updated"] == []

    @pytest.mark.anyio
    async def test_secrets_never_returned_in_config(self, client):
        await client.put(
            "/api/v1/config/secrets",
            json={"hf_token": "real-secret-value"},
        )

        resp = await client.get("/api/v1/config")
        config = resp.json()["config"]
        assert config["secrets"]["hf_token"] == SECRET_MASK
        assert "real-secret-value" not in str(config)
