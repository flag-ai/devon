from unittest.mock import MagicMock, patch
from collections import namedtuple

import pytest
from click.testing import CliRunner

from devon.cli.download_cmd import download
from devon.models.model_info import ModelMetadata


DiskUsage = namedtuple("DiskUsage", ["total", "used", "free"])


def _make_metadata():
    return ModelMetadata(
        source="huggingface",
        model_id="test/model",
        model_name="Test-Model",
        author="test",
        total_size_bytes=1000,
        file_count=3,
        parameter_count=7,
        format=["gguf"],
    )


@pytest.fixture
def runner():
    return CliRunner()


@pytest.fixture
def mock_env(tmp_path):
    """Mock all external dependencies of the download command."""
    with (
        patch("devon.cli.download_cmd.Settings") as mock_settings_cls,
        patch("devon.cli.download_cmd.ModelStorage") as mock_storage_cls,
        patch("devon.cli.download_cmd.SourceRegistry") as mock_registry,
        patch("devon.cli.download_cmd.shutil") as mock_shutil,
    ):
        settings = MagicMock()
        settings.storage_path = tmp_path / "models"
        mock_settings_cls.return_value = settings

        storage = MagicMock()
        storage.base_path = tmp_path / "models"
        storage.is_downloaded.return_value = False
        storage.get_model_path.return_value = tmp_path / "models" / "hf" / "test"
        mock_storage_cls.return_value = storage

        source = MagicMock()
        source.get_model_info.return_value = _make_metadata()
        source.download_model.return_value = ["model.gguf"]
        mock_registry.get_source.return_value = lambda: source

        mock_shutil.disk_usage.return_value = DiskUsage(
            total=500_000_000_000, used=100_000_000_000, free=400_000_000_000
        )

        yield {
            "source": source,
            "storage": storage,
            "settings": mock_settings_cls,
            "registry": mock_registry,
            "shutil": mock_shutil,
        }


class TestDownloadInclude:
    def test_include_forwarded_to_download_model(self, runner, mock_env):
        """--include patterns should be passed as allow_patterns to download_model."""
        runner.invoke(
            download,
            ["test/model", "--include", "*Q4_K_M*", "--include", "*Q5_K_M*", "-y"],
        )

        mock_env["source"].download_model.assert_called_once()
        call_kwargs = mock_env["source"].download_model.call_args
        assert call_kwargs.kwargs["allow_patterns"] == ["*Q4_K_M*", "*Q5_K_M*"]

    def test_no_include_passes_none(self, runner, mock_env):
        """Without --include, allow_patterns should be None."""
        runner.invoke(download, ["test/model", "-y"])

        mock_env["source"].download_model.assert_called_once()
        call_kwargs = mock_env["source"].download_model.call_args
        assert call_kwargs.kwargs["allow_patterns"] is None

    def test_yes_flag_skips_confirmation(self, runner, mock_env):
        """--yes flag should skip the confirmation prompt."""
        runner.invoke(download, ["test/model", "-y"])

        # Should not have prompted (no input needed) and should have called download
        mock_env["source"].download_model.assert_called_once()
