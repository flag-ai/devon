from devon.search.query_parser import QueryParser


class TestQueryParser:
    def test_empty_query(self):
        query, filters = QueryParser.parse("")
        assert query is None
        assert filters == {}

    def test_plain_text_query(self):
        query, filters = QueryParser.parse("llama chat")
        assert query == "llama chat"
        assert filters == {}

    def test_extract_params(self):
        query, filters = QueryParser.parse("qwen 30b")
        assert query == "qwen"
        assert filters["params"] == "30b"

    def test_extract_format(self):
        query, filters = QueryParser.parse("llama gguf")
        assert query == "llama"
        assert filters["format"] == "gguf"

    def test_extract_quantization(self):
        query, filters = QueryParser.parse("model Q4_K_M")
        assert query == "model"
        assert filters["quant"] == "Q4_K_M"

    def test_extract_multiple(self):
        query, filters = QueryParser.parse("llama 7b gguf")
        assert query == "llama"
        assert filters["params"] == "7b"
        assert filters["format"] == "gguf"

    def test_none_query(self):
        query, filters = QueryParser.parse(None)
        assert query is None
        assert filters == {}
