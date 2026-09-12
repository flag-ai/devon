import re
from typing import Optional, Tuple


class URLParser:
    """Parse model URLs to extract source and model ID."""

    PATTERNS = {
        "huggingface": [
            r"https?://(?:www\.)?huggingface\.co/([^/]+/[^/?]+)",
            r"https?://hf\.co/([^/]+/[^/?]+)",
        ],
    }

    @classmethod
    def parse(cls, url: str) -> Optional[Tuple[str, str]]:
        """
        Parse URL to extract source and model ID.

        Returns:
            Tuple of (source, model_id) or None
        """
        for source, patterns in cls.PATTERNS.items():
            for pattern in patterns:
                match = re.match(pattern, url)
                if match:
                    model_id = match.group(1).rstrip("/").split("?")[0]
                    return (source, model_id)
        return None

    @classmethod
    def is_url(cls, text: str) -> bool:
        """Check if text looks like a URL."""
        return text.startswith("http://") or text.startswith("https://")

    @classmethod
    def validate_url(cls, url: str) -> bool:
        """Check if URL is recognized."""
        return cls.parse(url) is not None
