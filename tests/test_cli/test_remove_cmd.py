from unittest.mock import MagicMock, patch

import pytest
from click.testing import CliRunner

from devon.cli.remove_cmd import remove


@pytest.fixture
def runner():
    return CliRunner()


@pytest.fixture
def mock_settings(tmp_path):
    with patch("devon.cli.remove_cmd.Settings") as mock:
        settings = MagicMock()
        settings.storage_path = tmp_path / "models"
        mock.return_value = settings
        yield mock


class TestRemoveCommand:
    def test_remove_happy_path(self, runner, mock_settings, tmp_path):
        """Remove an existing model with -y flag."""
        with patch("devon.cli.remove_cmd.ModelStorage") as mock_storage_cls:
            storage = MagicMock()
            storage.get_model_entry.return_value = {
                "source": "huggingface",
                "model_id": "test/model",
                "path": str(tmp_path / "models" / "huggingface" / "test" / "model"),
                "size_bytes": 5000,
            }
            storage.delete_model.return_value = True
            mock_storage_cls.return_value = storage

            result = runner.invoke(remove, ["test/model", "-y"])

            assert result.exit_code == 0
            assert "Removed" in result.output
            storage.delete_model.assert_called_once_with("huggingface", "test/model")

    def test_remove_missing_model(self, runner, mock_settings):
        """Removing a model not in the index should fail."""
        with patch("devon.cli.remove_cmd.ModelStorage") as mock_storage_cls:
            storage = MagicMock()
            storage.get_model_entry.return_value = None
            mock_storage_cls.return_value = storage

            result = runner.invoke(remove, ["missing/model", "-y"])

            assert result.exit_code != 0
            assert "not found" in result.output

    def test_remove_with_custom_source(self, runner, mock_settings, tmp_path):
        """Remove should use the specified source."""
        with patch("devon.cli.remove_cmd.ModelStorage") as mock_storage_cls:
            storage = MagicMock()
            storage.get_model_entry.return_value = {
                "source": "custom",
                "model_id": "test/model",
                "path": str(tmp_path / "models"),
                "size_bytes": 1000,
            }
            storage.delete_model.return_value = True
            mock_storage_cls.return_value = storage

            runner.invoke(remove, ["test/model", "--source", "custom", "-y"])

            storage.get_model_entry.assert_called_once_with("custom", "test/model")
            storage.delete_model.assert_called_once_with("custom", "test/model")

    def test_remove_confirmation_declined(self, runner, mock_settings):
        """Without -y, declining confirmation should not delete."""
        with patch("devon.cli.remove_cmd.ModelStorage") as mock_storage_cls:
            storage = MagicMock()
            storage.get_model_entry.return_value = {
                "source": "huggingface",
                "model_id": "test/model",
                "path": "/some/path",
                "size_bytes": 5000,
            }
            mock_storage_cls.return_value = storage

            runner.invoke(remove, ["test/model"], input="n\n")

            storage.delete_model.assert_not_called()
