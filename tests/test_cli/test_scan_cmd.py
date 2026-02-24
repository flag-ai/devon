from unittest.mock import MagicMock, patch

import pytest
from click.testing import CliRunner

from devon.cli.scan_cmd import scan


@pytest.fixture
def runner():
    return CliRunner()


@pytest.fixture
def mock_settings(tmp_path):
    with patch("devon.cli.scan_cmd.Settings") as mock:
        settings = MagicMock()
        settings.storage_path = tmp_path / "models"
        mock.return_value = settings
        yield mock


def _make_model_dir(base, source, model_id, extension=".safetensors"):
    """Helper to create a fake model directory on disk."""
    model_dir = base / source / model_id
    model_dir.mkdir(parents=True, exist_ok=True)
    (model_dir / f"model{extension}").write_bytes(b"x" * 1000)
    return model_dir


class TestScanCommand:
    def test_scan_discovers_new_model(self, runner, mock_settings, tmp_path):
        base = tmp_path / "models"
        _make_model_dir(base, "huggingface", "test/model")

        with patch("devon.cli.scan_cmd.ModelStorage") as mock_cls:
            storage = MagicMock()
            storage.base_path = base
            storage.index = {}
            mock_cls.return_value = storage

            result = runner.invoke(scan, [])

        assert result.exit_code == 0
        assert "1 added" in result.output

    def test_scan_dry_run(self, runner, mock_settings, tmp_path):
        base = tmp_path / "models"
        _make_model_dir(base, "huggingface", "test/model")

        with patch("devon.cli.scan_cmd.ModelStorage") as mock_cls:
            storage = MagicMock()
            storage.base_path = base
            storage.index = {}
            mock_cls.return_value = storage

            result = runner.invoke(scan, ["--dry-run"])

        assert result.exit_code == 0
        assert "dry run" in result.output
        storage._save_index.assert_not_called()

    def test_scan_reconcile_removes_stale(self, runner, mock_settings, tmp_path):
        base = tmp_path / "models"
        base.mkdir(parents=True)

        stale_entry = {
            "source": "huggingface",
            "model_id": "gone/model",
            "path": str(base / "huggingface" / "gone" / "model"),
            "metadata": {},
            "files": [],
            "downloaded_at": "2025-01-01T00:00:00",
            "last_used": None,
            "size_bytes": 500,
        }

        with patch("devon.cli.scan_cmd.ModelStorage") as mock_cls:
            storage = MagicMock()
            storage.base_path = base
            storage.index = {"huggingface::gone/model": stale_entry}
            mock_cls.return_value = storage

            result = runner.invoke(scan, ["--reconcile"])

        assert result.exit_code == 0
        assert "removed" in result.output

    def test_scan_stale_without_reconcile(self, runner, mock_settings, tmp_path):
        base = tmp_path / "models"
        base.mkdir(parents=True)

        stale_entry = {
            "source": "huggingface",
            "model_id": "gone/model",
            "path": str(base / "huggingface" / "gone" / "model"),
            "metadata": {},
            "files": [],
            "downloaded_at": "2025-01-01T00:00:00",
            "last_used": None,
            "size_bytes": 500,
        }

        with patch("devon.cli.scan_cmd.ModelStorage") as mock_cls:
            storage = MagicMock()
            storage.base_path = base
            storage.index = {"huggingface::gone/model": stale_entry}
            mock_cls.return_value = storage

            result = runner.invoke(scan, [])

        assert result.exit_code == 0
        assert "stale" in result.output

    def test_scan_custom_path(self, runner, mock_settings, tmp_path):
        base = tmp_path / "models"
        custom = tmp_path / "custom-models"
        _make_model_dir(custom, "local-stuff", "my-model", extension=".gguf")

        with patch("devon.cli.scan_cmd.ModelStorage") as mock_cls:
            storage = MagicMock()
            storage.base_path = base
            storage.index = {}
            mock_cls.return_value = storage

            result = runner.invoke(scan, [str(custom)])

        assert result.exit_code == 0
        assert "1 added" in result.output
