from pathlib import Path
from typing import List, Optional


# Known model file extensions and their format names
FORMAT_EXTENSIONS = {
    ".safetensors": "safetensors",
    ".bin": "pytorch",
    ".pt": "pytorch",
    ".pth": "pytorch",
    ".gguf": "gguf",
    ".ggml": "ggml",
    ".onnx": "onnx",
}


def detect_format(filepath: str) -> Optional[str]:
    """Detect model format from file extension."""
    path = Path(filepath)
    for ext, fmt in FORMAT_EXTENSIONS.items():
        if path.name.endswith(ext):
            return fmt
    return None


def detect_formats_from_files(filenames: List[str]) -> List[str]:
    """Detect all formats present in a list of files."""
    formats = set()
    for filename in filenames:
        fmt = detect_format(filename)
        if fmt:
            formats.add(fmt)
    return sorted(formats)
