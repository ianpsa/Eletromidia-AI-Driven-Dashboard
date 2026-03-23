from __future__ import annotations

from langchain_core.tools import tool


@tool
def filter_looker_dashboard(
    gender: str | None = None,
    age_range: str | None = None,
    city: str | None = None,
    social_class: list[str] | None = None,
) -> str:
    """Generate a filtered Looker Studio dashboard URL for the user to view.

    Use this tool when the user wants to visualize data or see a dashboard
    view of the campaign results.

    Args:
        gender: Filter by gender ('female' or 'male').
        age_range: Filter by age range (e.g. '20-29').
        city: Filter by city name.
        social_class: Filter by social classes (e.g. ['A', 'B1']).
    """
    # TODO: Implement Looker Studio URL generation.
    #
    # Expected implementation:
    #   1. Build the Looker Studio embed URL with filter params.
    #      Base: https://lookerstudio.google.com/embed/reporting/{REPORT_ID}/page/{PAGE_ID}
    #   2. Encode filters as URL config params per Looker Linking API.
    #   3. Return a JSON string with:
    #      - embed_url: the full URL for the frontend to render in an <iframe>
    #      - filters_applied: dict of filters that were applied
    #
    # Frontend will detect this tool's result and render an <iframe>.

    # TODO: mudar o frontend para detectar a chave "embed_url" na resposta e renderizar
    # um iframe com essa URL, ao invés de mostrar
    # o link diretamente para o usuário clicar.

    return "todo"
