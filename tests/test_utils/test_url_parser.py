from devon.utils.url_parser import URLParser


class TestURLParser:
    def test_parse_huggingface_url(self):
        result = URLParser.parse("https://huggingface.co/Qwen/Qwen2.5-32B-Instruct")
        assert result == ("huggingface", "Qwen/Qwen2.5-32B-Instruct")

    def test_parse_hf_short_url(self):
        result = URLParser.parse("https://hf.co/meta-llama/Llama-3.3-70B-Instruct")
        assert result == ("huggingface", "meta-llama/Llama-3.3-70B-Instruct")

    def test_parse_url_with_trailing_slash(self):
        result = URLParser.parse("https://huggingface.co/Qwen/Qwen2.5-32B-Instruct/")
        assert result == ("huggingface", "Qwen/Qwen2.5-32B-Instruct")

    def test_parse_url_with_query_params(self):
        result = URLParser.parse("https://huggingface.co/Qwen/Qwen2.5-32B-Instruct?ref=main")
        assert result == ("huggingface", "Qwen/Qwen2.5-32B-Instruct")

    def test_parse_unrecognized_url(self):
        result = URLParser.parse("https://example.com/model")
        assert result is None

    def test_is_url(self):
        assert URLParser.is_url("https://huggingface.co/model")
        assert URLParser.is_url("http://example.com")
        assert not URLParser.is_url("Qwen/Qwen2.5-32B")
        assert not URLParser.is_url("some-text")

    def test_validate_url(self):
        assert URLParser.validate_url("https://huggingface.co/Qwen/Qwen2.5-32B")
        assert not URLParser.validate_url("https://example.com/model")
