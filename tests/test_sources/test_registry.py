import pytest

from devon.sources.base import ModelSource
from devon.sources.registry import SourceRegistry, register_source


class TestSourceRegistry:
    def test_huggingface_registered(self):
        # HuggingFace should be auto-registered via import
        from devon.sources import HuggingFaceSource  # noqa: F401

        assert "huggingface" in SourceRegistry.list_all()

    def test_get_source(self):
        from devon.sources import HuggingFaceSource  # noqa: F401

        source_cls = SourceRegistry.get_source("huggingface")
        assert source_cls.name() == "huggingface"

    def test_get_missing_source(self):
        with pytest.raises(ValueError, match="not found"):
            SourceRegistry.get_source("nonexistent_source_xyz")

    def test_list_all(self):
        from devon.sources import HuggingFaceSource  # noqa: F401

        sources = SourceRegistry.list_all()
        assert isinstance(sources, list)
        assert "huggingface" in sources

    def test_register_custom_source(self):
        @register_source
        class TestSource(ModelSource):
            @classmethod
            def name(cls):
                return "test_source"

            @classmethod
            def is_available(cls):
                return True

            def search(self, query=None, filters=None, limit=20):
                return []

            def get_model_info(self, model_id):
                pass

            def download_model(self, model_id, dest_path, progress_callback=None):
                return []

        assert "test_source" in SourceRegistry.list_all()

        # Clean up
        del SourceRegistry._sources["test_source"]
