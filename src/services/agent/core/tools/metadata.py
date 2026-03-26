from __future__ import annotations

import logging

from langchain_core.tools import tool

from core.bigquery_client import get_dataset_ref, run_query

logger = logging.getLogger(__name__)


@tool
def get_available_filters() -> str:
    """Return the available filter options from the OOH dataset.

    Use this tool when the user asks what options exist or needs to understand
    what they can filter by (cities, age ranges, genders, social classes).
    """
    ds = get_dataset_ref()

    try:
        rows = run_query(
            f"""
            SELECT
              COUNT(DISTINCT CONCAT(endereco, ',', CAST(numero AS STRING)))
                AS total_points,
              ARRAY_AGG(DISTINCT cidade ORDER BY cidade LIMIT 500) AS cities
            FROM `{ds}.geodata`
            """
        )
    except Exception:
        logger.exception("Metadata filter query failed")
        return "Erro ao consultar filtros disponíveis. Tente novamente."

    total_points = rows[0]["total_points"] if rows else 0
    cities = rows[0]["cities"] if rows else []

    lines = [
        f"Total de pontos de mídia: {total_points}",
        "",
        f"Cidades disponíveis: {', '.join(cities)}",
        "Classes sociais: A, B1, B2, C1, C2, DE",
        "Faixas etárias: 18-19, 20-29, 30-39, 40-49, 50-59, 60-69, 70-79, 80+",
        "Gêneros: Feminino, Masculino",
    ]

    return "\n".join(lines)
