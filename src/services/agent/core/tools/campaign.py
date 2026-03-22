from __future__ import annotations

from typing import Optional

from google.cloud import bigquery
from langchain_core.tools import tool

from core.bigquery_client import get_dataset_ref, run_query_with_params

_AGE_BUCKETS: list[tuple[str, int, int]] = [
    ("x18_19", 18, 19),
    ("x20_29", 20, 29),
    ("x30_39", 30, 39),
    ("x40_49", 40, 49),
    ("x50_59", 50, 59),
    ("x60_69", 60, 69),
    ("x70_79", 70, 79),
    ("x80_plus", 80, 120),
]

_CLASS_COLUMNS: dict[str, str] = {
    "A": "a_class",
    "B1": "b1_class",
    "B2": "b2_class",
    "C1": "c1_class",
    "C2": "c2_class",
    "DE": "de_class",
}


def _overlapping_age_columns(age_min: int, age_max: int) -> list[str]:
    """Return BQ column names whose age range overlaps [age_min, age_max]."""
    return [col for col, lo, hi in _AGE_BUCKETS if lo <= age_max and hi >= age_min]


def _class_columns(classes: list[str]) -> list[str]:
    """Return BQ column names for the requested social classes."""
    cols = []
    for cls in classes:
        col = _CLASS_COLUMNS.get(cls.upper())
        if col:
            cols.append(col)
    return cols


def _build_sql(
    *,
    gender: Optional[str],
    age_min: Optional[int],
    age_max: Optional[int],
    classes: Optional[list[str]],
    city: Optional[str],
    latitude: Optional[float],
    longitude: Optional[float],
    radius_km: float,
    limit: int,
) -> tuple[str, list[bigquery.ScalarQueryParameter]]:
    """Build a BigQuery SQL query for campaign analysis.

    Returns (sql, params) for use with parameterized queries.
    """
    ds = get_dataset_ref()
    params: list[bigquery.ScalarQueryParameter] = []

    gender_expr = "1.0"
    if gender:
        gender_expr = "gnd.feminine" if gender.lower() == "female" else "gnd.masculine"

    age_expr = "1.0"
    if age_min is not None and age_max is not None:
        cols = _overlapping_age_columns(age_min, age_max)
        if cols:
            age_expr = " + ".join(f"a.{c}" for c in cols)

    class_expr = "1.0"
    if classes:
        cols = _class_columns(classes)
        if cols:
            class_expr = " + ".join(f"sc.{c}" for c in cols)

    where_parts: list[str] = []

    if latitude is not None and longitude is not None:
        radius_m = radius_km * 1000
        where_parts.append(
            "ST_DISTANCE("
            "ST_GEOGPOINT(CAST(g.longitude AS FLOAT64), "
            "CAST(g.latitude AS FLOAT64)), "
            "ST_GEOGPOINT(@lng, @lat)"
            ") <= @radius_m"
        )
        params.append(bigquery.ScalarQueryParameter("lng", "FLOAT64", longitude))
        params.append(bigquery.ScalarQueryParameter("lat", "FLOAT64", latitude))
        params.append(bigquery.ScalarQueryParameter("radius_m", "FLOAT64", radius_m))

    if city:
        where_parts.append("LOWER(g.cidade) = LOWER(@city)")
        params.append(bigquery.ScalarQueryParameter("city", "STRING", city))

    where_clause = ""
    if where_parts:
        where_clause = "WHERE " + " AND ".join(where_parts)

    sql = f"""
SELECT
  g.endereco,
  g.numero,
  ROUND(({gender_expr}) * ({age_expr}) * ({class_expr}) * 100, 2)
    AS affinity,
  ROUND(({gender_expr}) * ({age_expr}) * ({class_expr}) * g.uniques, 2)
    AS target_audience,
  ROUND(g.uniques, 2) AS total_flow
FROM `{ds}.geodata` g
JOIN `{ds}.target` t ON g.target_id = t.id
JOIN `{ds}.age` a ON t.age_id = a.id
JOIN `{ds}.gender` gnd ON t.gender_id = gnd.id
JOIN `{ds}.social_class` sc ON t.social_class_id = sc.id
{where_clause}
ORDER BY affinity DESC
LIMIT @result_limit
""".strip()

    params.append(bigquery.ScalarQueryParameter("result_limit", "INT64", limit))
    return sql, params


@tool
def analyze_campaign(
    gender: Optional[str] = None,
    age_min: Optional[int] = None,
    age_max: Optional[int] = None,
    classes: Optional[list[str]] = None,
    city: Optional[str] = None,
    latitude: Optional[float] = None,
    longitude: Optional[float] = None,
    radius_km: Optional[float] = 2.0,
    limit: Optional[int] = 5,
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
        limit: Maximum number of points to return (default 5).
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
        limit=limit or 5,
    )

    try:
        rows = run_query_with_params(sql, params)
    except Exception:
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
            f"{i}. {row['endereco']}, {row['numero']} — "
            f"Afinidade: {row['affinity']}%, "
            f"Público-alvo: {row['target_audience']}, "
            f"Fluxo total: {row['total_flow']}"
        )

    return "\n".join(lines)
