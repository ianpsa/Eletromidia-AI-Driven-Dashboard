"""Tests for pure functions in core.tools.looker."""

from __future__ import annotations

import json
from urllib.parse import unquote

from core.tools.looker import (
    _DEFAULT_PAGE_ID,
    _DEFAULT_REPORT_ID,
    _SEP,
    _build_url,
    _decode_unicode_escapes,
    _encode_value,
    _filter_eq,
    _filter_in,
    _get_config,
)


# ── _decode_unicode_escapes ───────────────────────────────────────────


class TestDecodeUnicodeEscapes:
    def test_basic_escape(self):
        assert _decode_unicode_escapes("\\u0041") == "A"

    def test_accented_char(self):
        assert _decode_unicode_escapes("caf\\u00e9") == "café"

    def test_no_escapes(self):
        assert _decode_unicode_escapes("hello") == "hello"

    def test_multiple_escapes(self):
        assert _decode_unicode_escapes("\\u0041\\u0042") == "AB"

    def test_empty_string(self):
        assert _decode_unicode_escapes("") == ""

    def test_uppercase_hex(self):
        assert _decode_unicode_escapes("\\u00E9") == "é"


# ── _encode_value ─────────────────────────────────────────────────────


class TestEncodeValue:
    def test_comma_replaced(self):
        assert _encode_value("-23.5505,-46.6333") == "-23.5505%2C-46.6333"

    def test_no_comma(self):
        assert _encode_value("São Paulo") == "São Paulo"

    def test_empty_string(self):
        assert _encode_value("") == ""

    def test_multiple_commas(self):
        assert _encode_value("a,b,c") == "a%2Cb%2Cc"

    def test_only_comma(self):
        assert _encode_value(",") == "%2C"


# ── _filter_eq ────────────────────────────────────────────────────────


class TestFilterEq:
    def test_structure(self):
        result = _filter_eq("São Paulo")
        parts = result.split(_SEP)
        assert parts[0] == "include"
        assert parts[1] == "0"
        assert parts[2] == "EQ"

    def test_value_included(self):
        result = _filter_eq("Rio de Janeiro")
        assert "Rio de Janeiro" in result

    def test_value_with_comma_encoded(self):
        # Coordinate values should have comma encoded
        result = _filter_eq("-23.5,-46.6")
        assert "%2C" in result
        assert ",-46.6" not in result

    def test_contains_sep(self):
        result = _filter_eq("value")
        assert _SEP in result


# ── _filter_in ────────────────────────────────────────────────────────


class TestFilterIn:
    def test_structure(self):
        result = _filter_in(["São Paulo"])
        parts = result.split(_SEP)
        assert parts[0] == "include"
        assert parts[1] == "0"
        assert parts[2] == "IN"

    def test_single_value(self):
        result = _filter_in(["São Paulo"])
        assert "São Paulo" in result

    def test_multiple_values_joined_by_sep(self):
        result = _filter_in(["São Paulo", "Rio de Janeiro"])
        assert "São Paulo" in result
        assert "Rio de Janeiro" in result
        assert _SEP in result

    def test_coordinates_encoded(self):
        result = _filter_in(["-23.5,-46.6"])
        assert "%2C" in result

    def test_empty_list(self):
        result = _filter_in([])
        # Should still have IN prefix
        assert "IN" in result


# ── _get_config ───────────────────────────────────────────────────────


class TestGetConfig:
    def test_defaults(self, monkeypatch):
        monkeypatch.delenv("LOOKER_REPORT_ID", raising=False)
        monkeypatch.delenv("LOOKER_PAGE_ID", raising=False)
        report_id, page_id = _get_config()
        assert report_id == _DEFAULT_REPORT_ID
        assert page_id == _DEFAULT_PAGE_ID

    def test_custom_report_id(self, monkeypatch):
        monkeypatch.setenv("LOOKER_REPORT_ID", "my-report-id")
        monkeypatch.delenv("LOOKER_PAGE_ID", raising=False)
        report_id, _ = _get_config()
        assert report_id == "my-report-id"

    def test_custom_page_id(self, monkeypatch):
        monkeypatch.delenv("LOOKER_REPORT_ID", raising=False)
        monkeypatch.setenv("LOOKER_PAGE_ID", "my-page-id")
        _, page_id = _get_config()
        assert page_id == "my-page-id"

    def test_both_custom(self, monkeypatch):
        monkeypatch.setenv("LOOKER_REPORT_ID", "rpt")
        monkeypatch.setenv("LOOKER_PAGE_ID", "pg")
        report_id, page_id = _get_config()
        assert report_id == "rpt"
        assert page_id == "pg"


# ── _build_url ────────────────────────────────────────────────────────


def _decode_params(url: str) -> dict:
    """Extract and JSON-decode the params= query parameter from a Looker URL."""
    _, params_encoded = url.split("params=", 1)
    return json.loads(unquote(params_encoded))


class TestBuildUrl:
    def test_no_filters_returns_base_url(self):
        url = _build_url(pontos=None, city=None, ambiente=None)
        assert "lookerstudio.google.com/embed/reporting" in url
        assert "params=" not in url

    def test_base_url_contains_report_and_page(self, monkeypatch):
        monkeypatch.delenv("LOOKER_REPORT_ID", raising=False)
        monkeypatch.delenv("LOOKER_PAGE_ID", raising=False)
        url = _build_url(pontos=None, city=None, ambiente=None)
        assert _DEFAULT_REPORT_ID in url
        assert _DEFAULT_PAGE_ID in url

    def test_city_filter_uses_eq(self, monkeypatch):
        monkeypatch.delenv("LOOKER_KEY_CIDADE", raising=False)
        url = _build_url(pontos=None, city="São Paulo", ambiente=None)
        data = _decode_params(url)
        assert "df49" in data
        assert "EQ" in data["df49"]
        assert "São Paulo" in data["df49"]

    def test_city_custom_key(self, monkeypatch):
        monkeypatch.setenv("LOOKER_KEY_CIDADE", "df77")
        url = _build_url(pontos=None, city="Campinas", ambiente=None)
        data = _decode_params(url)
        assert "df77" in data

    def test_pontos_filter_uses_in(self, monkeypatch):
        monkeypatch.delenv("LOOKER_KEY_PONTOS", raising=False)
        pontos = ["-23.5505,-46.6333", "-22.9068,-43.1729"]
        url = _build_url(pontos=pontos, city=None, ambiente=None)
        data = _decode_params(url)
        assert "df50" in data
        assert "IN" in data["df50"]

    def test_pontos_custom_key(self, monkeypatch):
        monkeypatch.setenv("LOOKER_KEY_PONTOS", "df99")
        url = _build_url(pontos=["-23.5,-46.6"], city=None, ambiente=None)
        data = _decode_params(url)
        assert "df99" in data

    def test_ambiente_filter(self, monkeypatch):
        monkeypatch.delenv("LOOKER_KEY_AMBIENTE", raising=False)
        url = _build_url(pontos=None, city=None, ambiente="Edifícios Residenciais")
        data = _decode_params(url)
        assert "df48" in data
        assert "EQ" in data["df48"]

    def test_combined_city_and_ambiente(self, monkeypatch):
        monkeypatch.delenv("LOOKER_KEY_CIDADE", raising=False)
        monkeypatch.delenv("LOOKER_KEY_AMBIENTE", raising=False)
        url = _build_url(pontos=None, city="São Paulo", ambiente="Outdoor")
        data = _decode_params(url)
        assert len(data) == 2
        assert "df49" in data
        assert "df48" in data

    def test_all_filters(self, monkeypatch):
        monkeypatch.delenv("LOOKER_KEY_PONTOS", raising=False)
        monkeypatch.delenv("LOOKER_KEY_CIDADE", raising=False)
        monkeypatch.delenv("LOOKER_KEY_AMBIENTE", raising=False)
        url = _build_url(
            pontos=["-23.5,-46.6"],
            city="São Paulo",
            ambiente="Outdoor",
        )
        data = _decode_params(url)
        assert len(data) == 3

    def test_unicode_in_city_decoded(self, monkeypatch):
        monkeypatch.delenv("LOOKER_KEY_CIDADE", raising=False)
        # LLM sometimes emits unicode escapes; they should be decoded
        url = _build_url(pontos=None, city="caf\\u00e9ville", ambiente=None)
        data = _decode_params(url)
        assert "caféville" in data["df49"]

    def test_coordinates_comma_encoded_in_pontos(self, monkeypatch):
        monkeypatch.delenv("LOOKER_KEY_PONTOS", raising=False)
        url = _build_url(pontos=["-23.5,-46.6"], city=None, ambiente=None)
        data = _decode_params(url)
        # The comma in coordinates must be encoded as %2C in the filter value
        assert "%2C" in data["df50"]
        # The literal comma must NOT appear (it would confuse the separator)
        value = data["df50"]
        # Remove known SEP occurrences; check no raw comma in the value segment
        after_in = value.split("IN" + _SEP, 1)[-1]
        assert "," not in after_in
