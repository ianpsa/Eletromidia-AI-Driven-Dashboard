from __future__ import annotations

import json
import os
import re
from urllib.parse import quote


def _decode_unicode_escapes(s: str) -> str:
    """Decode literal \\uXXXX sequences that LLMs sometimes emit."""
    return re.sub(r"\\u([0-9a-fA-F]{4})", lambda m: chr(int(m.group(1), 16)), s)

from langchain_core.tools import tool

_DEFAULT_REPORT_ID = "1776f716-b7de-4268-99ef-8107f950868d"
_DEFAULT_PAGE_ID = "DEQqF"
_DEFAULT_DS_ALIAS = "ds0"

# Looker Studio Linking API separator.
# The filter value inside the JSON must contain the literal string "%EE%80%80"
# (percent-encoded form of U+E000).  When quote() encodes the JSON string,
# these % signs get double-encoded to %25EE%2580%2580 — which is exactly the
# format Looker Studio expects in the URL.
_SEP = "%EE%80%80"


def _filter_value(value: str) -> str:
    """Encode a single value for the Looker Studio Linking API.

    Resulting format inside the JSON: include{SEP}0{SEP}IN{SEP}{url_encoded_value}
    Verified against real Looker Studio filter URLs.
    """
    return f"include{_SEP}0{_SEP}IN{_SEP}{quote(value, safe='')}"


def _get_config() -> tuple[str, str, str]:
    report_id = os.environ.get("LOOKER_REPORT_ID", _DEFAULT_REPORT_ID)
    page_id = os.environ.get("LOOKER_PAGE_ID", _DEFAULT_PAGE_ID)
    ds_alias = os.environ.get("LOOKER_DS_ALIAS", _DEFAULT_DS_ALIAS)
    return report_id, page_id, ds_alias


def _build_url(
    *,
    city: str | None,
    ambiente: str | None,
) -> str:
    """Build a Looker Studio embed URL with Linking API filter params."""
    report_id, page_id, ds_alias = _get_config()

    base = (
        f"https://lookerstudio.google.com/embed/reporting/{report_id}/page/{page_id}"
    )

    # Filter keys default to the field-name format "ds_alias.field".
    # If your report uses filter control IDs instead (df52, df53 …), override
    # via env vars LOOKER_KEY_CIDADE / LOOKER_KEY_AMBIENTE.
    filters: dict[str, str] = {}
    if city:
        key = os.environ.get("LOOKER_KEY_CIDADE", f"{ds_alias}.cidade")
        filters[key] = _filter_value(_decode_unicode_escapes(city))
    if ambiente:
        key = os.environ.get("LOOKER_KEY_AMBIENTE", f"{ds_alias}.ambiente")
        filters[key] = _filter_value(_decode_unicode_escapes(ambiente))

    if not filters:
        return base

    return f"{base}?params={quote(json.dumps(filters, separators=(',', ':'), ensure_ascii=False), safe=':,')}"


@tool
def filter_looker_dashboard(
    city: str | None = None,
    ambiente: str | None = None,
    gender: str | None = None,
    age_range: str | None = None,
    social_class: list[str] | None = None,
) -> str:
    """Generate a filtered Looker Studio dashboard URL for the user to view.

    Use this tool when the user wants to visualize data in a dashboard,
    see charts or maps of the campaign results, or asks for a visual view.

    The dashboard supports filtering by city and ambiente (screen location
    subtype).  Demographic filters (gender, age_range, social_class) are noted
    in the response but cannot be applied as visual filters — they are
    proportion columns, not categorical.

    Args:
        city: Filter by city name (e.g. 'São Paulo').
        ambiente: Filter by screen subtype, e.g. 'Edifícios Residenciais',
                  'Universidades', 'Hotéis'.
        gender: Demographic note — 'female' or 'male' (not a visual filter).
        age_range: Demographic note — e.g. '20-29' (not a visual filter).
        social_class: Demographic note — e.g. ['A', 'B1'] (not a visual filter).
    """
    url = _build_url(city=city, ambiente=ambiente)

    applied: list[str] = []
    if city:
        applied.append(f"cidade={city}")
    if ambiente:
        applied.append(f"ambiente={ambiente}")

    demo: list[str] = []
    if gender:
        label = "feminino" if gender.lower() == "female" else "masculino"
        demo.append(f"gênero={label}")
    if age_range:
        demo.append(f"idade={age_range}")
    if social_class:
        demo.append(f"classes={','.join(social_class)}")

    lines = [
        "Dashboard filtrado disponível:",
        url,
        "",
        f"Filtros aplicados no dashboard: {', '.join(applied) or 'nenhum'}",
    ]
    if demo:
        lines.append(
            f"Filtros demográficos (considerados na análise, não aplicáveis "
            f"como filtro visual): {', '.join(demo)}"
        )

    return "\n".join(lines)
