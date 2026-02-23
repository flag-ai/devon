"""Tests for static file serving and SPA fallback."""

import pytest
from httpx import ASGITransport, AsyncClient

from devon.config.settings import Settings
from devon.storage.organizer import ModelStorage


@pytest.fixture
def static_dir(tmp_path):
    """Create a fake static directory with test files."""
    static = tmp_path / "static"
    static.mkdir()

    (static / "index.html").write_text("<html><body>Devon UI</body></html>")

    assets = static / "assets"
    assets.mkdir()
    (assets / "index-abc123.js").write_text("console.log('app');")
    (assets / "index-abc123.css").write_text("body { color: red; }")

    return static


@pytest.fixture
def app(tmp_path, static_dir, monkeypatch):
    """Create a test app with static files available."""
    config_path = tmp_path / "config.yaml"
    storage_path = tmp_path / "models"
    storage_path.mkdir()

    monkeypatch.setenv("DEVON_API_KEY", "disable")
    monkeypatch.setattr("devon.api.app.STATIC_DIR", static_dir)

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


class TestSPAFallback:
    @pytest.mark.anyio
    async def test_root_serves_index_html(self, client):
        resp = await client.get("/")
        assert resp.status_code == 200
        assert "Devon UI" in resp.text

    @pytest.mark.anyio
    async def test_spa_route_falls_back_to_index(self, client):
        resp = await client.get("/search")
        assert resp.status_code == 200
        assert "Devon UI" in resp.text

    @pytest.mark.anyio
    async def test_nested_spa_route_falls_back(self, client):
        resp = await client.get("/models/some/deep/path")
        assert resp.status_code == 200
        assert "Devon UI" in resp.text

    @pytest.mark.anyio
    async def test_api_routes_not_intercepted(self, client):
        resp = await client.get("/api/v1/config")
        assert resp.status_code == 200
        data = resp.json()
        assert "config" in data

    @pytest.mark.anyio
    async def test_health_not_intercepted(self, client):
        resp = await client.get("/health")
        assert resp.status_code == 200
        assert resp.json()["status"] == "ok"


class TestPathTraversal:
    @pytest.mark.anyio
    async def test_traversal_returns_index(self, client):
        resp = await client.get("/../../etc/passwd")
        assert resp.status_code == 200
        assert "Devon UI" in resp.text

    @pytest.mark.anyio
    async def test_encoded_traversal_returns_index(self, client):
        resp = await client.get("/%2e%2e/%2e%2e/etc/passwd")
        assert resp.status_code == 200
        assert "Devon UI" in resp.text
