from __future__ import annotations

import os
import threading
from typing import Protocol

import httpx
from langchain_core.tools import tool

_DISAMBIGUATION_SUFFIX = ", município de São Paulo, Brasil"
_TIMEOUT = 10.0


class GeocodingProvider(Protocol):
    def geocode(self, query: str) -> tuple[float, float] | None: ...


class NominatimProvider:
    _BASE = "https://nominatim.openstreetmap.org/search"

    def geocode(self, query: str) -> tuple[float, float] | None:
        params = {
            "q": query + _DISAMBIGUATION_SUFFIX,
            "format": "json",
            "limit": "1",
            "countrycodes": "br",
        }
        headers = {"User-Agent": "EletromidiaAgent/1.0"}
        resp = httpx.get(self._BASE, params=params, headers=headers, timeout=_TIMEOUT)
        resp.raise_for_status()
        results = resp.json()
        if not results:
            return None
        return float(results[0]["lat"]), float(results[0]["lon"])


class MapboxProvider:
    _BASE = "https://api.mapbox.com/geocoding/v5/mapbox.places"

    def __init__(self) -> None:
        self._key = os.environ.get("MAPBOX_API_KEY", "")

    def geocode(self, query: str) -> tuple[float, float] | None:
        encoded = httpx.URL("", params={"q": query + _DISAMBIGUATION_SUFFIX}).params[
            "q"
        ]
        url = f"{self._BASE}/{encoded}.json"
        params = {"access_token": self._key, "limit": "1", "language": "pt"}
        resp = httpx.get(url, params=params, timeout=_TIMEOUT)
        resp.raise_for_status()
        features = resp.json().get("features", [])
        if not features:
            return None
        lon, lat = features[0]["center"]
        return float(lat), float(lon)


class GoogleProvider:
    _BASE = "https://maps.googleapis.com/maps/api/geocode/json"

    def __init__(self) -> None:
        self._key = os.environ.get("GOOGLE_MAPS_API_KEY", "")

    def geocode(self, query: str) -> tuple[float, float] | None:
        params = {"address": query + _DISAMBIGUATION_SUFFIX, "key": self._key}
        resp = httpx.get(self._BASE, params=params, timeout=_TIMEOUT)
        resp.raise_for_status()
        results = resp.json().get("results", [])
        if not results:
            return None
        loc = results[0]["geometry"]["location"]
        return float(loc["lat"]), float(loc["lng"])


_PROVIDERS: dict[str, type[GeocodingProvider]] = {
    "nominatim": NominatimProvider,
    "mapbox": MapboxProvider,
    "google": GoogleProvider,
}

_cached_provider: GeocodingProvider | None = None
_provider_lock = threading.Lock()


def get_provider() -> GeocodingProvider:
    global _cached_provider
    if _cached_provider is not None:
        return _cached_provider
    with _provider_lock:
        if _cached_provider is not None:
            return _cached_provider
        name = os.environ.get("GEOCODING_PROVIDER", "nominatim").lower()
        cls = _PROVIDERS.get(name)
        if cls is None:
            valid = ", ".join(_PROVIDERS)
            raise ValueError(
                f"Unknown GEOCODING_PROVIDER={name!r}. Choose from: {valid}"
            )
        _cached_provider = cls()
        return _cached_provider


@tool
def geocode_location(query: str) -> str:
    """Convert a location name or address into geographic coordinates.

    Use this tool when the user mentions a specific place, neighborhood,
    street, or landmark and you need latitude/longitude to filter media
    points by geographic radius.

    Always use full street names (e.g. 'Avenida Brigadeiro Faria Lima'
    instead of 'Faria Lima').

    Args:
        query: Location name, address, or landmark to geocode.
    """
    provider = get_provider()
    result = provider.geocode(query)
    if result is None:
        return f"Localização não encontrada: {query}"
    lat, lon = result
    return f"Coordenadas encontradas: latitude={lat}, longitude={lon}"
