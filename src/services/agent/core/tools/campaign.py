from __future__ import annotations

import logging

from google.cloud import bigquery
from langchain_core.tools import tool

from core.bigquery_client import get_dataset_ref, run_query_with_params

logger = logging.getLogger(__name__)

_AGE_BUCKETS: list[tuple[str, int, int]] = [
    ("p_18_19", 18, 19),
    ("p_20_29", 20, 29),
    ("p_30_39", 30, 39),
    ("p_40_49", 40, 49),
    ("p_50_59", 50, 59),
    ("p_60_69", 60, 69),
    ("p_70_79", 70, 79),
    ("p_80_plus", 80, 120),
]

_CLASS_COLUMNS: dict[str, str] = {
    "A": "p_a",
    "B1": "p_b1",
    "B2": "p_b2",
    "C1": "p_c1",
    "C2": "p_c2",
    "DE": "p_de",
}


def _overlapping_age_columns(age_min: int, age_max: int) -> list[str]:
    """Return enriched_screens column names whose age range overlaps [age_min, age_max]."""
    return [col for col, lo, hi in _AGE_BUCKETS if lo <= age_max and hi >= age_min]


def _class_columns(classes: list[str]) -> list[str]:
    """Return enriched_screens column names for the requested social classes."""
    cols = []
    for cls in classes:
        col = _CLASS_COLUMNS.get(cls.upper())
        if col:
            cols.append(col)
    return cols


def _build_sql(
    *,
    gender: str | None,
    age_min: int | None,
    age_max: int | None,
    classes: list[str] | None,
    city: str | None,
    vertical: str | None,
    ambiente: str | None,
    latitude: float | None,
    longitude: float | None,
    radius_km: float,
    limit: int,
) -> tuple[str, list[bigquery.ScalarQueryParameter]]:
    """Build a BigQuery SQL query against the enriched_screens table.

    The enriched_screens table is denormalised: each row is one Eletromidia
    screen with pre-computed demographic proportions from the Claro spatial
    join.  The affinity score is the product of the relevant proportion
    columns.
    """
    ds = get_dataset_ref()
    params: list[bigquery.ScalarQueryParameter] = []

    gender_expr = "1.0"
    if gender:
        gender_expr = "s.p_f" if gender.lower() == "female" else "s.p_m"

    age_expr = "1.0"
    if age_min is not None and age_max is not None:
        cols = _overlapping_age_columns(age_min, age_max)
        if cols:
            age_expr = " + ".join(f"s.{c}" for c in cols)

    class_expr = "1.0"
    if classes:
        cols = _class_columns(classes)
        if cols:
            class_expr = " + ".join(f"s.{c}" for c in cols)

    where_parts: list[str] = []

    if latitude is not None and longitude is not None:
        radius_m = radius_km * 1000
        where_parts.append(
            "ST_DISTANCE("
            "ST_GEOGPOINT(s.longitude, s.latitude), "
            "ST_GEOGPOINT(@lng, @lat)"
            ") <= @radius_m"
        )
        params.append(bigquery.ScalarQueryParameter("lng", "FLOAT64", longitude))
        params.append(bigquery.ScalarQueryParameter("lat", "FLOAT64", latitude))
        params.append(bigquery.ScalarQueryParameter("radius_m", "FLOAT64", radius_m))

    if city:
        where_parts.append("LOWER(s.cidade) = LOWER(@city)")
        params.append(bigquery.ScalarQueryParameter("city", "STRING", city))

    if vertical:
        where_parts.append("LOWER(s.vertical) = LOWER(@vertical)")
        params.append(bigquery.ScalarQueryParameter("vertical", "STRING", vertical))

    if ambiente:
        where_parts.append("LOWER(s.ambiente) = LOWER(@ambiente)")
        params.append(bigquery.ScalarQueryParameter("ambiente", "STRING", ambiente))

    where_clause = ""
    if where_parts:
        where_clause = "WHERE " + " AND ".join(where_parts)

    sql = f"""
SELECT
  s.endereco_ref,
  s.vertical,
  s.ambiente,
  s.cidade,
  ROUND(({gender_expr}) * ({age_expr}) * ({class_expr}) * 100, 2)
    AS affinity,
  ROUND(({gender_expr}) * ({age_expr}) * ({class_expr}) * s.uniques, 2)
    AS target_audience,
  ROUND(s.uniques, 2) AS total_flow,
  s.match_type
FROM `{ds}.enriched_screens` s
{where_clause}
ORDER BY affinity DESC
LIMIT @result_limit
""".strip()

    params.append(bigquery.ScalarQueryParameter("result_limit", "INT64", limit))
    return sql, params


@tool
def analyze_campaign(
    gender: str | None = None,
    age_min: int | None = None,
    age_max: int | None = None,
    classes: list[str] | None = None,
    city: str | None = None,
    vertical: str | None = None,
    ambiente: str | None = None,
    latitude: float | None = None,
    longitude: float | None = None,
    radius_km: float | None = 2.0,
    limit: int | None = 5,
) -> str:
    """Analyze OOH media screens and return a ranked list by audience affinity.

    Use this tool when the user asks for campaign recommendations, best media
    points, or audience targeting.

    If you have a latitude/longitude from geocoding, pass them to filter screens
    by geographic radius.  Otherwise you can pass a city name for a broad filter.

    Args:
        gender: Target gender — 'female' or 'male'.
        age_min: Minimum target age.
        age_max: Maximum target age.
        classes: Target social classes, e.g. ['A', 'B1'].
        city: City name for broad filtering.
        vertical: Screen location type — 'Edifícios', 'MUB-Rua',
                  'Estabelecimentos Comerciais', or 'Shoppings'.
        ambiente: Screen location subtype, e.g. 'Edifícios Residenciais',
                  'Shoppings Experiência', 'Universidades', 'Hotéis', etc.
        latitude: Center latitude for geographic filtering.
        longitude: Center longitude for geographic filtering.
        radius_km: Radius in km around the center point (default 2).
        limit: Maximum number of screens to return (default 5).
    """
    sql, params = _build_sql(
        gender=gender,
        age_min=age_min,
        age_max=age_max,
        classes=classes,
        city=city,
        vertical=vertical,
        ambiente=ambiente,
        latitude=latitude,
        longitude=longitude,
        radius_km=radius_km or 2.0,
        limit=limit or 5,
    )

    try:
        rows = run_query_with_params(sql, params)
    except Exception:
        logger.exception("Campaign query failed")
        return "Erro ao consultar BigQuery. Verifique os filtros e tente novamente."

    if not rows:
        return "Nenhum ponto de mídia encontrado para os filtros informados."

    filters_desc: list[str] = []
    if gender:
        filters_desc.append(f"gênero={gender}")
    if age_min is not None or age_max is not None:
        filters_desc.append(f"idade={age_min or '?'}-{age_max or '?'}")
    if classes:
        filters_desc.append(f"classes={','.join(classes)}")
    if city:
        filters_desc.append(f"cidade={city}")
    if vertical:
        filters_desc.append(f"vertical={vertical}")
    if ambiente:
        filters_desc.append(f"ambiente={ambiente}")
    if latitude is not None and longitude is not None:
        filters_desc.append(
            f"raio={radius_km or 2.0}km ({latitude:.4f}, {longitude:.4f})"
        )

    lines = [
        f"Resultados: {len(rows)} telas.",
        f"Filtros: {'; '.join(filters_desc) or 'nenhum'}",
        "",
        "Ranking:",
    ]
    for i, row in enumerate(rows, 1):
        lines.append(
            f"{i}. {row['endereco_ref']} ({row['vertical']} — {row['ambiente']}) — "
            f"Afinidade: {row['affinity']}%, "
            f"Público-alvo: {row['target_audience']}, "
            f"Fluxo total: {row['total_flow']}"
        )

    return "\n".join(lines)
