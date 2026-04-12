from __future__ import annotations

import logging
import re

from langchain_core.tools import tool

from core.bigquery_client import get_dataset_ref, run_query

logger = logging.getLogger(__name__)

_LIMIT_RE = re.compile(r"\bLIMIT\s+\d+", re.IGNORECASE)
_DANGEROUS_RE = re.compile(
    r"\b(?:INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|TRUNCATE|MERGE"
    r"|GRANT|REVOKE|CALL|EXEC|EXECUTE)\b",
    re.IGNORECASE,
)
_SELECT_ONLY_ERROR = "Erro: apenas consultas SELECT são permitidas."


def _validate_sql(sql: str) -> str | None:
    """Return an error string if *sql* is not a safe read-only query, else None."""
    if ";" in sql:
        return _SELECT_ONLY_ERROR

    normalized = sql.upper().lstrip()
    if not (normalized.startswith("SELECT") or normalized.startswith("WITH")):
        return _SELECT_ONLY_ERROR
    if _DANGEROUS_RE.search(normalized):
        return _SELECT_ONLY_ERROR

    return None


@tool
def query_bigquery(sql_query: str) -> str:
    """Execute a SQL query against the BigQuery OOH dataset.

    Use this tool when the user asks for custom data analysis, aggregations,
    or questions that cannot be answered by the pre-built campaign tool.

    Tables (fully qualify with the dataset automatically):
      - geodata: id, impression_hour (INT64, hour of day 0-23 — NOT a date),
                 location_id, uniques,
                 latitude (STRING), longitude (STRING),
                 uf_estado, cidade, endereco, numero, target_id
                 NOTE: geodata has NO date column; do NOT filter by date here.
      - target: id, age_id, gender_id, social_class_id
      - age: id, x18_19, x20_29, x30_39, x40_49, x50_59,
             x60_69, x70_79, x80_plus  (FLOAT percentages)
      - gender: id, feminine, masculine  (FLOAT percentages)
      - social_class: id, a_class, b1_class, b2_class,
                      c1_class, c2_class, de_class  (FLOAT percentages)

    Args:
        sql_query: A BigQuery SQL query (SELECT only). Table names without
                   project/dataset prefix will be auto-qualified.
    """
    dataset = get_dataset_ref()

    sql = sql_query.strip().rstrip(";")

    error = _validate_sql(sql)
    if error:
        return error

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
        logger.exception("BigQuery query failed")
        return "Erro ao executar query. Verifique a sintaxe SQL e tente novamente."

    if not rows:
        return "Nenhum resultado encontrado."

    headers = list(rows[0].keys())
    lines = [" | ".join(headers)]
    lines.append("-" * len(lines[0]))
    for row in rows:
        lines.append(" | ".join(str(row.get(h, "")) for h in headers))

    return f"Resultados ({len(rows)} linhas):\n" + "\n".join(lines)
