from __future__ import annotations

import json
import os
import re
from typing import Optional
from urllib.parse import quote

from langchain_core.tools import tool

_DEFAULT_REPORT_ID = "1776f716-b7de-4268-99ef-8107f950868d"
_DEFAULT_PAGE_ID = "p_dmgnzqj61d"

# Literal separator string used inside the JSON filter values.
# Looker Studio expects this exact string (percent-encoded form of U+E000)
# as delimiter between filter parts. It gets double-encoded when the whole
# JSON is URL-encoded: %EE%80%80 → %25EE%2580%2580 in the final URL.
_SEP = "%EE%80%80"


def _decode_unicode_escapes(s: str) -> str:
    """Decode literal \\uXXXX sequences that LLMs sometimes emit."""
    return re.sub(r"\\u([0-9a-fA-F]{4})", lambda m: chr(int(m.group(1), 16)), s)


def _encode_value(v: str) -> str:
    """Encode a coordinate/value: replace comma with %2C."""
    return v.replace(",", "%2C")


def _filter_eq(value: str) -> str:
    """Single-value EQ filter string (placed inside JSON value)."""
    return f"include{_SEP}0{_SEP}EQ{_SEP}{_encode_value(value)}"


def _filter_in(values: list[str]) -> str:
    """Multi-value IN filter string (placed inside JSON value)."""
    encoded = _SEP.join(_encode_value(v) for v in values)
    return f"include{_SEP}0{_SEP}IN{_SEP}{encoded}"


def _get_config() -> tuple[str, str]:
    report_id = os.environ.get("LOOKER_REPORT_ID", _DEFAULT_REPORT_ID)
    page_id = os.environ.get("LOOKER_PAGE_ID", _DEFAULT_PAGE_ID)
    return report_id, page_id


def _build_url(
    *,
    pontos: Optional[list[str]],
    city: Optional[str],
    ambiente: Optional[str],
) -> str:
    """Build a Looker Studio embed URL with df filter control params."""
    report_id, page_id = _get_config()
    base = f"https://lookerstudio.google.com/embed/reporting/{report_id}/page/{page_id}"

    filters: dict[str, str] = {}

    if pontos:
        key = os.environ.get("LOOKER_KEY_PONTOS", "df50")
        decoded = [_decode_unicode_escapes(p) for p in pontos]
        filters[key] = _filter_in(decoded)

    if city:
        key = os.environ.get("LOOKER_KEY_CIDADE", "df49")
        filters[key] = _filter_eq(_decode_unicode_escapes(city))

    if ambiente:
        key = os.environ.get("LOOKER_KEY_AMBIENTE", "df48")
        filters[key] = _filter_eq(_decode_unicode_escapes(ambiente))

    if not filters:
        return base

    params_json = json.dumps(filters, separators=(",", ":"))
    return f"{base}?params={quote(params_json, safe=':')}"


@tool
def filter_looker_dashboard(
    pontos: Optional[list[str]] = None,
    city: Optional[str] = None,
    ambiente: Optional[str] = None,
) -> str:
    """Generate a Looker Studio dashboard URL filtered to specific OOH points.

    Call this tool AFTER analyze_campaign to show the returned points in the
    Looker dashboard.  Extract the coordinates from each result line
    (format: [coords: lat,lng]) and pass them as the pontos list.

    Args:
        pontos: List of geographic coordinates as "lat,lng" strings (e.g.
                ["-23.5505,-46.6333", "-23.5489,-46.6388"]).  These map to
                the ooh_pontos dimension in Looker.
        city: Optional city name to also filter by cidade.
        ambiente: Optional screen subtype filter (e.g. 'Edifícios Residenciais').
    """
    url = _build_url(pontos=pontos, city=city, ambiente=ambiente)

    applied: list[str] = []
    if pontos:
        applied.append(f"{len(pontos)} pontos selecionados")
    if city:
        applied.append(f"cidade={city}")
    if ambiente:
        applied.append(f"ambiente={ambiente}")

    lines = [
        "Dashboard com os pontos recomendados:",
        url,
        "",
        f"Filtros aplicados: {', '.join(applied) or 'nenhum'}",
    ]

    return "\n".join(lines)
