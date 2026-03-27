from __future__ import annotations

import logging

from langchain_core.tools import tool

from core.bigquery_client import get_dataset_ref, run_query

logger = logging.getLogger(__name__)


@tool
def get_available_filters() -> str:
    """Return the available filter options from the OOH dataset.

    Use this tool when the user asks what options exist or needs to understand
    what they can filter by (cities, verticals, ambientes, age ranges, genders,
    social classes).
    """
    ds = get_dataset_ref()

    try:
        rows = run_query(
            f"""
            SELECT
              COUNT(*) AS total_screens,
              ARRAY_AGG(DISTINCT cidade ORDER BY cidade LIMIT 500) AS cities,
              ARRAY_AGG(DISTINCT vertical ORDER BY vertical LIMIT 50) AS verticals,
              ARRAY_AGG(DISTINCT ambiente ORDER BY ambiente LIMIT 100) AS ambientes
            FROM `{ds}.enriched_screens`
            """
        )
    except Exception:
        logger.exception("Metadata filter query failed")
        return "Erro ao consultar filtros disponíveis. Tente novamente."

    total = rows[0]["total_screens"] if rows else 0
    cities = rows[0]["cities"] if rows else []
    verticals = rows[0]["verticals"] if rows else []
    ambientes = rows[0]["ambientes"] if rows else []

    lines = [
        f"Total de telas de mídia: {total}",
        "",
        f"Cidades disponíveis: {', '.join(str(c) for c in cities)}",
        f"Verticais (tipo de local): {', '.join(str(v) for v in verticals)}",
        f"Ambientes (subtipo): {', '.join(str(a) for a in ambientes)}",
        "Classes sociais: A, B1, B2, C1, C2, DE",
        "Faixas etárias: 18-19, 20-29, 30-39, 40-49, 50-59, 60-69, 70-79, 80+",
        "Gêneros: Feminino, Masculino",
    ]

    return "\n".join(lines)
