import os

import pytest
import yaml

from devon.config.settings import DEFAULT_CONFIG, SECRET_MASK, Settings


@pytest.fixture
def settings_path(tmp_path):
    """Return a config path in a temp directory (file does not exist yet)."""
    return tmp_path / "config" / "config.yaml"


@pytest.fixture
def settings(settings_path):
    """Create a Settings instance with no existing config file."""
    return Settings(config_path=settings_path)


class TestIsConfigured:
    def test_false_when_no_file(self, settings, settings_path):
        assert not settings_path.exists()
        assert settings.is_configured is False

    def test_true_after_save(self, settings, settings_path):
        settings.save()
        assert settings_path.exists()
        assert settings.is_configured is True


class TestUpdate:
    def test_merges_and_persists(self, settings, settings_path):
        settings.update({"storage": {"base_path": "/custom/path"}})

        assert settings.get("storage.base_path") == "/custom/path"
        # Other defaults preserved
        assert settings.get("download.resume") is True

        # Verify on-disk persistence
        assert settings_path.exists()
        with open(settings_path) as f:
            on_disk = yaml.safe_load(f)
        assert on_disk["storage"]["base_path"] == "/custom/path"

    def test_deep_merge_preserves_siblings(self, settings):
        settings.update({"storage": {"max_size_gb": 100}})

        # base_path should still be the default
        assert settings.get("storage.base_path") == DEFAULT_CONFIG["storage"]["base_path"]
        assert settings.get("storage.max_size_gb") == 100

    def test_sets_hf_token_env_var(self, settings, monkeypatch):
        monkeypatch.delenv("HF_TOKEN", raising=False)
        settings.update({"secrets": {"hf_token": "test-token-value"}})
        assert os.environ.get("HF_TOKEN") == "test-token-value"

    def test_does_not_set_env_when_token_none(self, settings, monkeypatch):
        monkeypatch.delenv("HF_TOKEN", raising=False)
        settings.update({"storage": {"base_path": "/foo"}})
        assert os.environ.get("HF_TOKEN") is None


class TestToSafeDict:
    def test_masks_set_secrets(self, settings):
        settings.update({"secrets": {"hf_token": "real-token", "api_key": "real-key"}})

        safe = settings.to_safe_dict()
        assert safe["secrets"]["hf_token"] == SECRET_MASK
        assert safe["secrets"]["api_key"] == SECRET_MASK

    def test_does_not_mask_none_secrets(self, settings):
        safe = settings.to_safe_dict()
        assert safe["secrets"]["hf_token"] is None
        assert safe["secrets"]["api_key"] is None

    def test_does_not_mutate_internal_config(self, settings):
        settings.update({"secrets": {"hf_token": "real-token"}})

        safe = settings.to_safe_dict()
        assert safe["secrets"]["hf_token"] == SECRET_MASK
        # Internal config should still have the real value
        assert settings.get("secrets.hf_token") == "real-token"

    def test_includes_all_sections(self, settings):
        safe = settings.to_safe_dict()
        for section in DEFAULT_CONFIG:
            assert section in safe
