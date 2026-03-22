from __future__ import annotations

import os

from langchain_core.messages import (
    BaseMessage,
    HumanMessage,
    RemoveMessage,
    SystemMessage,
    get_buffer_string,
)
from langgraph.graph import MessagesState

from core.llm_provider import get_llm_provider

_CONTEXT_WINDOWS: dict[str, int] = {
    # Groq
    "llama-3.3-70b-versatile": 128_000,
    "llama-3.1-70b-versatile": 128_000,
    # Google
    "gemini-2.0-flash": 1_048_576,
}
_DEFAULT_CONTEXT = 128_000
_TRIGGER_RATIO = 0.75
_KEEP_RECENT = 6

SUMMARIZATION_PROMPT = (
    "Sua tarefa é criar um resumo detalhado da conversa abaixo entre um usuário "
    "e um agente de planejamento de mídia OOH (Out-of-Home) da Eletromidia.\n\n"
    "Este resumo será usado como contexto para continuar a conversa sem perder "
    "informações críticas. Seja minucioso em capturar detalhes técnicos e decisões "
    "que seriam essenciais para continuar o atendimento sem perda de contexto.\n\n"
    "O resumo DEVE incluir as seguintes seções:\n\n"
    "1. Objetivo e Intenção do Usuário:\n"
    "   Capture todos os pedidos explícitos do usuário em detalhe. Qual é o "
    "objetivo da campanha? Qual público-alvo desejado? Há restrições de "
    "orçamento, prazo ou região?\n\n"
    "2. Filtros e Parâmetros de Segmentação:\n"
    "   Liste todos os filtros discutidos e aplicados:\n"
    "   - Gênero (feminino/masculino)\n"
    "   - Faixa etária (ex: 25-34)\n"
    "   - Classes sociais (ex: A, B1)\n"
    "   - Cidade ou região\n"
    "   - Coordenadas geográficas e raio de busca\n"
    "   Inclua tanto os filtros finais quanto os que foram alterados ao longo "
    "da conversa.\n\n"
    "3. Ferramentas Utilizadas e Resultados:\n"
    "   Para cada ferramenta chamada, registre:\n"
    "   - Qual ferramenta foi usada (analyze_campaign, query_bigquery, "
    "geocode_location, filter_looker_dashboard, get_available_filters)\n"
    "   - Quais parâmetros foram passados\n"
    "   - Resumo dos resultados: pontos de mídia recomendados (com endereços "
    "e números), valores de afinidade, público-alvo estimado, fluxo total.\n"
    "   Preserve nomes de ruas, números, coordenadas e valores numéricos "
    "exatos — estes são críticos.\n\n"
    "4. Correções e Feedback do Usuário:\n"
    "   Registre TODAS as correções, preferências alteradas ou sugestões "
    "rejeitadas pelo usuário. Preste atenção especial a feedback como "
    "'não era isso', 'mude para', 'prefiro X em vez de Y'. Estas "
    "informações são fundamentais para não repetir erros.\n\n"
    "5. Todas as Mensagens do Usuário:\n"
    "   Liste TODAS as mensagens do usuário (não as respostas de ferramentas). "
    "Estas são essenciais para entender a evolução da intenção e mudanças "
    "de direção.\n\n"
    "6. Solicitações Pendentes:\n"
    "   Liste qualquer pedido que foi feito mas ainda não foi atendido ou "
    "concluído. Inclua o contexto de por que está pendente.\n\n"
    "7. Estado Atual da Conversa:\n"
    "   Descreva com precisão o que estava sendo discutido imediatamente "
    "antes deste ponto. Qual foi a última ação tomada? O que o usuário "
    "espera como próximo passo?\n\n"
    "Responda apenas com o resumo estruturado, sem explicações adicionais. "
    "Use os títulos das seções acima para organizar."
)


def _estimate_tokens(messages: list[BaseMessage]) -> int:
    """Rough token estimate: chars / 4."""
    total = 0
    for msg in messages:
        content = msg.content
        if isinstance(content, str):
            total += len(content) // 4
        elif isinstance(content, list):
            for block in content:
                if isinstance(block, dict):
                    total += len(block.get("text", "")) // 4
                elif isinstance(block, str):
                    total += len(block) // 4
    return total


def _get_context_window() -> int:
    """Look up context window for the configured model."""
    model = os.environ.get("LLM_MODEL", "")
    return _CONTEXT_WINDOWS.get(model, _DEFAULT_CONTEXT)


def maybe_summarize(state: MessagesState) -> dict:
    """Graph node: summarize old messages if approaching context window limit."""
    messages = state["messages"]

    context_window = _get_context_window()
    threshold = int(context_window * _TRIGGER_RATIO)
    estimated = _estimate_tokens(messages)

    if estimated <= threshold:
        return {"messages": []}

    keep = _KEEP_RECENT
    old_messages = messages[:-keep] if keep < len(messages) else []

    if not old_messages:
        return {"messages": []}

    llm = get_llm_provider().build()
    summary_input = [
        SystemMessage(content=SUMMARIZATION_PROMPT),
        HumanMessage(content=get_buffer_string(old_messages)),
    ]
    summary_response = llm.invoke(summary_input)

    removals = [RemoveMessage(id=m.id) for m in old_messages]
    summary_msg = SystemMessage(
        content=f"[Resumo da conversa anterior]\n{summary_response.content}"
    )
    return {"messages": removals + [summary_msg]}
