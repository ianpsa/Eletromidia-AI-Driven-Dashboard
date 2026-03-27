from __future__ import annotations

import logging

from google.cloud import bigquery
from langchain_core.tools import tool

from core.bigquery_client import get_dataset_ref, run_query_with_params

logger = logging.getLogger(__name__)

_AGE_BUCKETS: list[tuple[str, int, int]] = [
    ("age_18_19_count", 18, 19),
    ("age_20_29_count", 20, 29),
    ("age_30_39_count", 30, 39),
    ("age_40_49_count", 40, 49),
    ("age_50_59_count", 50, 59),
    ("age_60_69_count", 60, 69),
    ("age_70_79_count", 70, 79),
    ("age_80_plus_count", 80, 120),
]

_CLASS_COLUMNS: dict[str, str] = {
    "A": "class_a_count",
    "B1": "class_b1_count",
    "B2": "class_b2_count",
    "C1": "class_c1_count",
    "C2": "class_c2_count",
    "DE": "class_de_count",
}


def _overlapping_age_columns(age_min: int, age_max: int) -> list[str]:
    return [col for col, lo, hi in _AGE_BUCKETS if lo <= age_max and hi >= age_min]


def _class_columns(classes: list[str]) -> list[str]:
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
    latitude: float | None,
    longitude: float | None,
    radius_km: float,
    limit: int,
) -> tuple[str, list[bigquery.ScalarQueryParameter]]:
    """Build a BigQuery SQL query against vw_geodata_enriched.

    The view has demographic counts per location and impression_hour.
    We aggregate by location, compute proportions from counts, and rank by
    the product of the relevant demographic proportions (affinity score).
    """
    ds = get_dataset_ref()
    params: list[bigquery.ScalarQueryParameter] = []

    # Affinity numerator: sum of relevant demographic counts
    # Affinity = (target_count / total_uniques) * 100
    # We build each dimension as a SUM expression over the count columns.

    gender_sum = "SUM(s.feminine_count)" if gender and gender.lower() == "female" else (
        "SUM(s.masculine_count)" if gender and gender.lower() == "male" else "SUM(s.uniques)"
    )

    if age_min is not None and age_max is not None:
        cols = _overlapping_age_columns(age_min, age_max)
        age_sum = " + ".join(f"SUM(s.{c})" for c in cols) if cols else "SUM(s.uniques)"
    else:
        age_sum = "SUM(s.uniques)"

    if classes:
        cols = _class_columns(classes)
        class_sum = " + ".join(f"SUM(s.{c})" for c in cols) if cols else "SUM(s.uniques)"
    else:
        class_sum = "SUM(s.uniques)"

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
        where_parts.append("s.latitude IS NOT NULL")

    if city:
        where_parts.append("LOWER(s.cidade) = LOWER(@city)")
        params.append(bigquery.ScalarQueryParameter("city", "STRING", city))

    where_clause = ("WHERE " + " AND ".join(where_parts)) if where_parts else ""

    sql = f"""
SELECT
  s.location_id,
  CONCAT(TRIM(s.endereco), ', ', CAST(s.numero AS STRING), ' — ', s.cidade) AS endereco_ref,
  s.cidade,
  ANY_VALUE(s.latitude) AS latitude,
  ANY_VALUE(s.longitude) AS longitude,
  ROUND(SUM(s.uniques), 0) AS total_flow,
  ROUND(
    (({gender_sum}) / NULLIF(SUM(s.uniques), 0)) *
    (({age_sum}) / NULLIF(SUM(s.uniques), 0)) *
    (({class_sum}) / NULLIF(SUM(s.uniques), 0)) * 100,
    2
  ) AS affinity,
  ROUND(
    SUM(s.uniques) *
    (({gender_sum}) / NULLIF(SUM(s.uniques), 0)) *
    (({age_sum}) / NULLIF(SUM(s.uniques), 0)) *
    (({class_sum}) / NULLIF(SUM(s.uniques), 0)),
    0
  ) AS target_audience
FROM `{ds}.vw_geodata_enriched` s
{where_clause}
GROUP BY s.location_id, s.endereco, s.numero, s.cidade
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
    latitude: float | None = None,
    longitude: float | None = None,
    radius_km: float | None = 2.0,
    limit: int | None = 10,
) -> str:
    """Analyze OOH media points and return a ranked list by audience affinity.

    Use this tool when the user asks for campaign recommendations, best media
    points, or audience targeting.

    If you have a latitude/longitude from geocoding, pass them to filter points
    by geographic radius.  Otherwise you can pass a city name for a broad filter.

    Args:
        gender: Target gender — 'female' or 'male'.
        age_min: Minimum target age.
        age_max: Maximum target age.
        classes: Target social classes, e.g. ['A', 'B1'].
        city: City name for broad filtering.
        latitude: Center latitude for geographic filtering.
        longitude: Center longitude for geographic filtering.
        radius_km: Radius in km around the center point (default 2).
        limit: Maximum number of points to return (default 10).
    """
    sql, params = _build_sql(
        gender=gender,
        age_min=age_min,
        age_max=age_max,
        classes=classes,
        city=city,
        latitude=latitude,
        longitude=longitude,
        radius_km=radius_km or 2.0,
        limit=limit or 10,
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
    if latitude is not None and longitude is not None:
        filters_desc.append(
            f"raio={radius_km or 2.0}km ({latitude:.4f}, {longitude:.4f})"
        )

    lines = [
        f"Resultados: {len(rows)} pontos.",
        f"Filtros: {'; '.join(filters_desc) or 'nenhum'}",
        "",
        "Ranking:",
    ]
    for i, row in enumerate(rows, 1):
        lines.append(
            f"{i}. {row['endereco_ref']} — "
            f"Afinidade: {row['affinity']}%, "
            f"Público-alvo: {int(row['target_audience'] or 0)}, "
            f"Fluxo total: {int(row['total_flow'] or 0)}"
        )

    return "\n".join(lines)
