from __future__ import annotations

import re

from langchain_core.tools import tool

from core.bigquery_client import get_dataset_ref, run_query

_LIMIT_RE = re.compile(r"\bLIMIT\s+\d+", re.IGNORECASE)


@tool
def query_bigquery(sql_query: str) -> str:
    """Execute a SQL query against the BigQuery OOH dataset.

    Use this tool when the user asks for custom data analysis, aggregations,
    or questions that cannot be answered by the pre-built campaign tool.

    Tables (fully qualify with the dataset automatically):
      - geodata: id, impression_hour, location_id, uniques,
                 latitude (STRING), longitude (STRING),
                 uf_estado, cidade, endereco, numero, target_id
      - target: id, age_id, gender_id, social_class_id
      - age: id, x18_19, x20_29, x30_39, x40_49, x50_59,
             x60_69, x70_79, x80_plus  (FLOAT percentages)
      - gender: id, feminine, masculine  (FLOAT percentages)
      - social_class: id, a_class, b1_class, b2_class,
                      c1_class, c2_class, de_class  (FLOAT percentages)

    Geographic: use ST_DISTANCE(ST_GEOGPOINT(CAST(longitude AS FLOAT64),
    CAST(latitude AS FLOAT64)), ST_GEOGPOINT(lon, lat)) for distance.

    Args:
        sql_query: A BigQuery SQL query. Table names without project/dataset
                   prefix will be auto-qualified.
    """
    dataset = get_dataset_ref()

    sql = sql_query.strip().rstrip(";")

    normalized = sql.lstrip("(")
    if not normalized.upper().startswith("SELECT"):
        return "Erro: apenas consultas SELECT são permitidas."

    tables = ["geodata", "target", "age", "gender", "social_class"]
    for table in tables:
        sql = re.sub(
            rf"\b{table}\b(?!\.)(?!`)",
            f"`{dataset}.{table}`",
            sql,
        )

    if not _LIMIT_RE.search(sql):
        sql += " LIMIT 100"

    try:
        rows = run_query(sql)
    except Exception:
        return "Erro ao executar query. Verifique a sintaxe SQL e tente novamente."

    if not rows:
        return "Nenhum resultado encontrado."

    headers = list(rows[0].keys())
    lines = [" | ".join(headers)]
    lines.append("-" * len(lines[0]))
    for row in rows:
        lines.append(" | ".join(str(row.get(h, "")) for h in headers))

    return f"Resultados ({len(rows)} linhas):\n" + "\n".join(lines)
