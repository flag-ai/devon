from devon.utils.format_utils import detect_format, detect_formats_from_files


class TestDetectFormat:
    def test_safetensors(self):
        assert detect_format("model.safetensors") == "safetensors"

    def test_pytorch_bin(self):
        assert detect_format("pytorch_model.bin") == "pytorch"

    def test_pytorch_pt(self):
        assert detect_format("model.pt") == "pytorch"

    def test_gguf(self):
        assert detect_format("model-Q4_K_M.gguf") == "gguf"

    def test_onnx(self):
        assert detect_format("model.onnx") == "onnx"

    def test_unknown(self):
        assert detect_format("README.md") is None

    def test_with_path(self):
        assert detect_format("models/llama/model.safetensors") == "safetensors"


class TestDetectFormatsFromFiles:
    def test_multiple_formats(self):
        files = ["model.safetensors", "model.gguf", "README.md"]
        result = detect_formats_from_files(files)
        assert "safetensors" in result
        assert "gguf" in result
        assert len(result) == 2

    def test_no_models(self):
        files = ["README.md", "config.json"]
        result = detect_formats_from_files(files)
        assert result == []

    def test_empty(self):
        result = detect_formats_from_files([])
        assert result == []
