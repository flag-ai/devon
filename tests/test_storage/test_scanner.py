import json

import pytest

from devon.storage.scanner import ModelScanner


@pytest.fixture
def scanner():
    return ModelScanner()


@pytest.fixture
def model_tree(tmp_path):
    """Create a model directory tree with various model types."""
    base = tmp_path / "models"

    # HuggingFace model with config.json
    hf_model = base / "huggingface" / "Qwen" / "Qwen2.5-7B-Instruct"
    hf_model.mkdir(parents=True)
    (hf_model / "model.safetensors").write_bytes(b"x" * 2000)
    (hf_model / "config.json").write_text(
        json.dumps(
            {
                "model_type": "qwen2",
                "hidden_size": 3584,
            }
        )
    )

    # GGUF model (local/custom)
    gguf_model = base / "my-models" / "llama-7B-Q4_K_M"
    gguf_model.mkdir(parents=True)
    (gguf_model / "llama-7B-Q4_K_M.gguf").write_bytes(b"x" * 3000)

    # Model with no known source
    custom = base / "experiments" / "custom-bert"
    custom.mkdir(parents=True)
    (custom / "model.bin").write_bytes(b"x" * 500)
    (custom / "config.json").write_text(json.dumps({"model_type": "bert"}))

    return base


class TestModelScanner:
    def test_scan_finds_new_models(self, scanner, model_tree):
        results = scanner.scan(model_tree, existing_keys=set())
        assert len(results) == 3

    def test_scan_skips_existing(self, scanner, model_tree):
        results = scanner.scan(
            model_tree,
            existing_keys={"huggingface::Qwen/Qwen2.5-7B-Instruct"},
        )
        assert len(results) == 2
        model_ids = {r["model_id"] for r in results}
        assert "Qwen/Qwen2.5-7B-Instruct" not in model_ids

    def test_scan_infers_hf_source(self, scanner, model_tree):
        results = scanner.scan(model_tree, existing_keys=set())
        hf = [r for r in results if r["source"] == "huggingface"]
        assert len(hf) == 1
        assert hf[0]["model_id"] == "Qwen/Qwen2.5-7B-Instruct"

    def test_scan_infers_local_source(self, scanner, model_tree):
        results = scanner.scan(model_tree, existing_keys=set())
        local = [r for r in results if r["source"] == "local"]
        assert len(local) == 2

    def test_scan_infers_architecture(self, scanner, model_tree):
        results = scanner.scan(model_tree, existing_keys=set())
        hf = [r for r in results if r["model_id"] == "Qwen/Qwen2.5-7B-Instruct"][0]
        assert hf["metadata"]["architecture"] == "qwen"

        bert = [r for r in results if "custom-bert" in r["model_id"]][0]
        assert bert["metadata"]["architecture"] == "bert"

    def test_scan_infers_quantization(self, scanner, model_tree):
        results = scanner.scan(model_tree, existing_keys=set())
        gguf = [r for r in results if "Q4_K_M" in r["model_id"]][0]
        assert gguf["metadata"]["quantization"] == "Q4_K_M"

    def test_scan_infers_format(self, scanner, model_tree):
        results = scanner.scan(model_tree, existing_keys=set())
        hf = [r for r in results if r["model_id"] == "Qwen/Qwen2.5-7B-Instruct"][0]
        assert "safetensors" in hf["metadata"]["format"]

        gguf = [r for r in results if "Q4_K_M" in r["model_id"]][0]
        assert "gguf" in gguf["metadata"]["format"]

    def test_scan_infers_parameter_count(self, scanner, model_tree):
        results = scanner.scan(model_tree, existing_keys=set())
        hf = [r for r in results if r["model_id"] == "Qwen/Qwen2.5-7B-Instruct"][0]
        assert hf["metadata"]["parameter_count"] == 7

        gguf = [r for r in results if "llama-7B" in r["model_id"]][0]
        assert gguf["metadata"]["parameter_count"] == 7

    def test_scan_calculates_size(self, scanner, model_tree):
        results = scanner.scan(model_tree, existing_keys=set())
        hf = [r for r in results if r["model_id"] == "Qwen/Qwen2.5-7B-Instruct"][0]
        # 2000 bytes for safetensors + config.json size
        assert hf["size_bytes"] > 2000

    def test_scan_infers_author(self, scanner, model_tree):
        results = scanner.scan(model_tree, existing_keys=set())
        hf = [r for r in results if r["model_id"] == "Qwen/Qwen2.5-7B-Instruct"][0]
        assert hf["metadata"]["author"] == "Qwen"
        assert hf["metadata"]["model_name"] == "Qwen2.5-7B-Instruct"

    def test_scan_empty_dir(self, scanner, tmp_path):
        base = tmp_path / "empty"
        base.mkdir()
        results = scanner.scan(base, existing_keys=set())
        assert results == []

    def test_scan_nonexistent_dir(self, scanner, tmp_path):
        results = scanner.scan(tmp_path / "nope", existing_keys=set())
        assert results == []


class TestFindStale:
    def test_find_stale_detects_missing(self, scanner, tmp_path):
        manifest = {
            "huggingface::gone/model": {
                "source": "huggingface",
                "model_id": "gone/model",
                "path": str(tmp_path / "does-not-exist"),
                "size_bytes": 100,
            },
        }
        stale = scanner.find_stale(manifest)
        assert stale == ["huggingface::gone/model"]

    def test_find_stale_keeps_existing(self, scanner, tmp_path):
        model_dir = tmp_path / "models" / "huggingface" / "test" / "model"
        model_dir.mkdir(parents=True)
        (model_dir / "model.bin").write_bytes(b"data")

        manifest = {
            "huggingface::test/model": {
                "source": "huggingface",
                "model_id": "test/model",
                "path": str(model_dir),
                "size_bytes": 4,
            },
        }
        stale = scanner.find_stale(manifest)
        assert stale == []

    def test_find_stale_empty_manifest(self, scanner):
        assert scanner.find_stale({}) == []
