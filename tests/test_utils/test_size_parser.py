from devon.utils.size_parser import format_bytes, format_number, parse_params, parse_size


class TestParseSize:
    def test_parse_gb(self):
        assert parse_size("100gb") == 100 * 1024**3

    def test_parse_mb(self):
        assert parse_size("500mb") == 500 * 1024**2

    def test_parse_tb(self):
        assert parse_size("1tb") == 1024**4

    def test_parse_with_spaces(self):
        assert parse_size("  100 gb  ") == 100 * 1024**3

    def test_parse_invalid(self):
        assert parse_size("invalid") is None

    def test_parse_decimal(self):
        assert parse_size("1.5gb") == int(1.5 * 1024**3)


class TestFormatBytes:
    def test_bytes(self):
        assert format_bytes(500) == "500.0B"

    def test_kb(self):
        assert format_bytes(1024) == "1.0KB"

    def test_mb(self):
        assert format_bytes(1024**2) == "1.0MB"

    def test_gb(self):
        assert format_bytes(1024**3) == "1.0GB"

    def test_tb(self):
        assert format_bytes(1024**4) == "1.0TB"

    def test_negative(self):
        assert format_bytes(-1) == "0B"


class TestFormatNumber:
    def test_plain(self):
        assert format_number(500) == "500"

    def test_thousands(self):
        assert format_number(1500) == "1.5K"

    def test_millions(self):
        assert format_number(2_500_000) == "2.5M"


class TestParseParams:
    def test_with_b(self):
        assert parse_params("30b") == 30

    def test_without_b(self):
        assert parse_params("7") == 7

    def test_uppercase(self):
        assert parse_params("70B") == 70

    def test_invalid(self):
        assert parse_params("abc") is None
