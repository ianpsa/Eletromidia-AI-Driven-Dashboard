from __future__ import annotations

import json
import os
from urllib.parse import quote

from langchain_core.tools import tool

_DEFAULT_REPORT_ID = "1776f716-b7de-4268-99ef-8107f950868d"
_DEFAULT_PAGE_ID = "p0"
_DEFAULT_DS_ALIAS = "ds0"


def _get_config() -> tuple[str, str, str]:
    report_id = os.environ.get("LOOKER_REPORT_ID", _DEFAULT_REPORT_ID)
    page_id = os.environ.get("LOOKER_PAGE_ID", _DEFAULT_PAGE_ID)
    ds_alias = os.environ.get("LOOKER_DS_ALIAS", _DEFAULT_DS_ALIAS)
    return report_id, page_id, ds_alias


def _build_url(
    *,
    city: str | None,
    vertical: str | None,
    ambiente: str | None,
) -> str:
    """Build a Looker Studio embed URL with Linking API filter params."""
    report_id, page_id, ds_alias = _get_config()

    base = (
        f"https://lookerstudio.google.com/embed/reporting/{report_id}/page/{page_id}"
    )

    filters: dict[str, str] = {}
    if city:
        filters[f"{ds_alias}.cidade"] = f"include\x00{city}"
    if vertical:
        filters[f"{ds_alias}.vertical"] = f"include\x00{vertical}"
    if ambiente:
        filters[f"{ds_alias}.ambiente"] = f"include\x00{ambiente}"

    if not filters:
        return base

    params_json = json.dumps(filters, ensure_ascii=False)
    params_json = params_json.replace("\\u0000", "\x00")
    return f"{base}?params={quote(params_json)}"


@tool
def filter_looker_dashboard(
    city: str | None = None,
    vertical: str | None = None,
    ambiente: str | None = None,
    gender: str | None = None,
    age_range: str | None = None,
    social_class: list[str] | None = None,
) -> str:
    """Generate a filtered Looker Studio dashboard URL for the user to view.

    Use this tool when the user wants to visualize data in a dashboard,
    see charts or maps of the campaign results, or asks for a visual view.

    The dashboard supports filtering by city, vertical (screen location type),
    and ambiente (screen location subtype).  Demographic filters (gender,
    age_range, social_class) are noted in the response but cannot be applied
    as visual filters — they are proportion columns, not categorical.

    Args:
        city: Filter by city name (e.g. 'São Paulo').
        vertical: Filter by screen type — 'Edifícios', 'MUB-Rua',
                  'Estabelecimentos Comerciais', or 'Shoppings'.
        ambiente: Filter by screen subtype, e.g. 'Edifícios Residenciais',
                  'Universidades', 'Hotéis'.
        gender: Demographic note — 'female' or 'male' (not a visual filter).
        age_range: Demographic note — e.g. '20-29' (not a visual filter).
        social_class: Demographic note — e.g. ['A', 'B1'] (not a visual filter).
    """
    url = _build_url(city=city, vertical=vertical, ambiente=ambiente)

    applied: list[str] = []
    if city:
        applied.append(f"cidade={city}")
    if vertical:
        applied.append(f"vertical={vertical}")
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
