import re
from typing import Optional


def parse_size(size_str: str) -> Optional[int]:
    """
    Parse human-readable size string to bytes.

    Examples:
        "100gb" -> 107374182400
        "500mb" -> 524288000
        "1.5tb" -> 1649267441664
    """
    match = re.match(r"(\d+(?:\.\d+)?)\s*(b|kb|mb|gb|tb|pb)", size_str.lower().strip())
    if not match:
        return None

    value = float(match.group(1))
    unit = match.group(2)

    multipliers = {
        "b": 1,
        "kb": 1024,
        "mb": 1024**2,
        "gb": 1024**3,
        "tb": 1024**4,
        "pb": 1024**5,
    }

    # int() truncates sub-byte fractions, which is correct — bytes are whole numbers
    return int(value * multipliers[unit])


def format_bytes(bytes_val: float) -> str:
    """Format bytes to human-readable string."""
    if bytes_val < 0:
        return "0B"
    for unit in ["B", "KB", "MB", "GB", "TB"]:
        if bytes_val < 1024.0:
            return f"{bytes_val:.1f}{unit}"
        bytes_val /= 1024.0
    return f"{bytes_val:.1f}PB"


def parse_params(params_str: str) -> Optional[int]:
    """
    Parse parameter count string.

    Examples:
        "30b" -> 30
        "7B" -> 7
        "70" -> 70
    """
    match = re.match(r"(\d+)\s*b\b", params_str.lower().strip())
    if not match:
        match = re.match(r"(\d+)$", params_str.lower().strip())
    return int(match.group(1)) if match else None


def format_number(num: int) -> str:
    """Format numbers with K/M suffixes."""
    if num >= 1_000_000:
        return f"{num / 1_000_000:.1f}M"
    elif num >= 1_000:
        return f"{num / 1_000:.1f}K"
    return str(num)
