"""Tests for core.tools.geocoding — provider selection and HTTP responses."""

from __future__ import annotations

from unittest.mock import MagicMock, patch

import pytest

import core.tools.geocoding as geo_module
from core.tools.geocoding import (
    GoogleProvider,
    MapboxProvider,
    NominatimProvider,
    get_provider,
)

# ── helpers ───────────────────────────────────────────────────────────


def _mock_http(json_body: object) -> MagicMock:
    resp = MagicMock()
    resp.json.return_value = json_body
    resp.raise_for_status.return_value = None
    return resp


# ── NominatimProvider ─────────────────────────────────────────────────


class TestNominatimProvider:
    def test_returns_lat_lon(self):
        body = [{"lat": "-23.5505", "lon": "-46.6333"}]
        with patch("core.tools.geocoding.httpx.get", return_value=_mock_http(body)):
            result = NominatimProvider().geocode("Av. Paulista")
        assert result == (-23.5505, -46.6333)

    def test_returns_none_on_empty(self):
        with patch("core.tools.geocoding.httpx.get", return_value=_mock_http([])):
            result = NominatimProvider().geocode("Endereço inexistente")
        assert result is None

    def test_query_appends_disambiguation_suffix(self):
        with patch(
            "core.tools.geocoding.httpx.get", return_value=_mock_http([])
        ) as mock_get:
            NominatimProvider().geocode("Faria Lima")
        params = mock_get.call_args[1]["params"]
        assert "município de São Paulo, Brasil" in params["q"]
        assert "Faria Lima" in params["q"]

    def test_uses_countrycodes_br(self):
        with patch(
            "core.tools.geocoding.httpx.get", return_value=_mock_http([])
        ) as mock_get:
            NominatimProvider().geocode("Rua Augusta")
        params = mock_get.call_args[1]["params"]
        assert params["countrycodes"] == "br"

    def test_result_is_float_tuple(self):
        body = [{"lat": "-22.9068", "lon": "-43.1729"}]
        with patch("core.tools.geocoding.httpx.get", return_value=_mock_http(body)):
            result = NominatimProvider().geocode("Rio de Janeiro")
        assert isinstance(result, tuple)
        assert all(isinstance(v, float) for v in result)


# ── MapboxProvider ────────────────────────────────────────────────────


class TestMapboxProvider:
    def test_returns_lat_lon(self, monkeypatch):
        monkeypatch.setenv("MAPBOX_API_KEY", "test-key")
        body = {"features": [{"center": [-46.6333, -23.5505]}]}
        with patch("core.tools.geocoding.httpx.get", return_value=_mock_http(body)):
            result = MapboxProvider().geocode("Av. Paulista")
        # Mapbox returns [lon, lat]; provider should flip to (lat, lon)
        assert result == (-23.5505, -46.6333)

    def test_returns_none_on_empty_features(self, monkeypatch):
        monkeypatch.setenv("MAPBOX_API_KEY", "test-key")
        body = {"features": []}
        with patch("core.tools.geocoding.httpx.get", return_value=_mock_http(body)):
            result = MapboxProvider().geocode("Endereço inexistente")
        assert result is None

    def test_result_is_float_tuple(self, monkeypatch):
        monkeypatch.setenv("MAPBOX_API_KEY", "test-key")
        body = {"features": [{"center": [-43.1729, -22.9068]}]}
        with patch("core.tools.geocoding.httpx.get", return_value=_mock_http(body)):
            result = MapboxProvider().geocode("Rio")
        assert isinstance(result, tuple)
        assert all(isinstance(v, float) for v in result)


# ── GoogleProvider ────────────────────────────────────────────────────


class TestGoogleProvider:
    def test_returns_lat_lon(self, monkeypatch):
        monkeypatch.setenv("GOOGLE_MAPS_API_KEY", "test-key")
        body = {
            "results": [
                {"geometry": {"location": {"lat": -23.5505, "lng": -46.6333}}}
            ]
        }
        with patch("core.tools.geocoding.httpx.get", return_value=_mock_http(body)):
            result = GoogleProvider().geocode("Av. Paulista")
        assert result == (-23.5505, -46.6333)

    def test_returns_none_on_empty_results(self, monkeypatch):
        monkeypatch.setenv("GOOGLE_MAPS_API_KEY", "test-key")
        body = {"results": []}
        with patch("core.tools.geocoding.httpx.get", return_value=_mock_http(body)):
            result = GoogleProvider().geocode("Endereço inexistente")
        assert result is None

    def test_result_is_float_tuple(self, monkeypatch):
        monkeypatch.setenv("GOOGLE_MAPS_API_KEY", "test-key")
        body = {
            "results": [
                {"geometry": {"location": {"lat": -22.9068, "lng": -43.1729}}}
            ]
        }
        with patch("core.tools.geocoding.httpx.get", return_value=_mock_http(body)):
            result = GoogleProvider().geocode("Rio de Janeiro")
        assert isinstance(result, tuple)
        assert all(isinstance(v, float) for v in result)


# ── get_provider ──────────────────────────────────────────────────────


@pytest.fixture(autouse=True)
def reset_cached_provider():
    """Reset the module-level cached provider before and after each test."""
    geo_module._cached_provider = None
    yield
    geo_module._cached_provider = None


class TestGetProvider:
    def test_default_is_nominatim(self, monkeypatch):
        monkeypatch.delenv("GEOCODING_PROVIDER", raising=False)
        assert isinstance(get_provider(), NominatimProvider)

    def test_nominatim_explicit(self, monkeypatch):
        monkeypatch.setenv("GEOCODING_PROVIDER", "nominatim")
        assert isinstance(get_provider(), NominatimProvider)

    def test_mapbox_provider(self, monkeypatch):
        monkeypatch.setenv("GEOCODING_PROVIDER", "mapbox")
        monkeypatch.setenv("MAPBOX_API_KEY", "key")
        assert isinstance(get_provider(), MapboxProvider)

    def test_google_provider(self, monkeypatch):
        monkeypatch.setenv("GEOCODING_PROVIDER", "google")
        monkeypatch.setenv("GOOGLE_MAPS_API_KEY", "key")
        assert isinstance(get_provider(), GoogleProvider)

    def test_unknown_provider_raises_value_error(self, monkeypatch):
        monkeypatch.setenv("GEOCODING_PROVIDER", "unknown_xyz")
        with pytest.raises(ValueError, match="Unknown GEOCODING_PROVIDER"):
            get_provider()

    def test_provider_is_cached_on_second_call(self, monkeypatch):
        monkeypatch.delenv("GEOCODING_PROVIDER", raising=False)
        p1 = get_provider()
        p2 = get_provider()
        assert p1 is p2

    def test_provider_name_case_insensitive(self, monkeypatch):
        monkeypatch.setenv("GEOCODING_PROVIDER", "NOMINATIM")
        assert isinstance(get_provider(), NominatimProvider)
