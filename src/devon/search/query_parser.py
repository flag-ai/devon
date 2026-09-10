import re
from typing import Any, Dict, Optional, Tuple


class QueryParser:
    """Parse user search queries into structured filters."""

    # Patterns for inline filters in query strings
    FILTER_PATTERNS = {
        "params": re.compile(r"\b(\d+)b\b", re.IGNORECASE),
        "format": re.compile(r"\b(gguf|safetensors|pytorch|onnx)\b", re.IGNORECASE),
        "quant": re.compile(r"\b(Q4_K_M|Q5_K_M|Q8_0|fp16|bf16|int8|int4)\b"),
    }

    @classmethod
    def parse(cls, query: str) -> Tuple[Optional[str], Dict[str, Any]]:
        """
        Parse a query string, extracting inline filters.

        Args:
            query: Raw query string like "qwen 30b gguf"

        Returns:
            Tuple of (cleaned_query, extracted_filters)
        """
        if not query:
            return None, {}

        filters = {}
        clean_query = query

        # Extract parameter count
        match = cls.FILTER_PATTERNS["params"].search(query)
        if match:
            filters["params"] = match.group(0)
            clean_query = clean_query.replace(match.group(0), "").strip()

        # Extract format
        match = cls.FILTER_PATTERNS["format"].search(query)
        if match:
            filters["format"] = match.group(1).lower()
            clean_query = clean_query.replace(match.group(0), "").strip()

        # Extract quantization
        match = cls.FILTER_PATTERNS["quant"].search(query)
        if match:
            filters["quant"] = match.group(1)
            clean_query = clean_query.replace(match.group(0), "").strip()

        # Clean up extra whitespace
        clean_query = " ".join(clean_query.split())

        return clean_query or None, filters
