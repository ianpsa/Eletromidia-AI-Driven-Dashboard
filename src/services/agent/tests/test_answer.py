import json
from unittest.mock import MagicMock, call, patch

import numpy as np
import pytest

from core.answer import generate_final_answer


@patch("core.answer.ChatGroq")
def test_retorna_content_da_resposta(mock_groq_cls):
    """Deve retornar o conteúdo textual da resposta da LLM."""
    mock_llm = MagicMock()
    mock_llm.invoke.return_value.content = "Recomendo os pontos X e Y."
    mock_groq_cls.return_value = mock_llm

    result = generate_final_answer(
        user_prompt="campanha feminina",
        filters={"gender": "female"},
        ranking=[{"point": "Av. A, 1", "total_target": 100, "total_flow": 500, "affinity": 20.0}],
        api_key="fake-key",
    )

    assert result == "Recomendo os pontos X e Y."


@patch("core.answer.ChatGroq")
def test_contexto_contem_campos_esperados(mock_groq_cls):
    """A instrução enviada à LLM deve conter os campos de contexto esperados."""
    mock_llm = MagicMock()
    mock_llm.invoke.return_value.content = "resposta"
    mock_groq_cls.return_value = mock_llm

    generate_final_answer(
        user_prompt="campanha classe A",
        filters={"classes": ["A"]},
        ranking=[{"point": "Av. B, 2", "total_target": 50, "total_flow": 200, "affinity": 25.0}],
        api_key="fake-key",
        city_fallback=True,
        used_age_range=(18, 35),
    )

    instruction = mock_llm.invoke.call_args[0][0]
    assert "campanha classe A" in instruction
    assert "pergunta_usuario" in instruction
    assert "filtros_aplicados" in instruction
    assert "cidade_nao_encontrada" in instruction
    assert "faixa_utilizada" in instruction
    assert "top_pontos" in instruction


@patch("core.answer.ChatGroq")
def test_ranking_truncado_a_cinco(mock_groq_cls):
    """O ranking no contexto deve ser limitado a 5 itens."""
    mock_llm = MagicMock()
    mock_llm.invoke.return_value.content = "resposta"
    mock_groq_cls.return_value = mock_llm

    ranking = [
        {"point": f"Ponto {i}", "total_target": 10, "total_flow": 100, "affinity": 10.0}
        for i in range(10)
    ]

    generate_final_answer(
        user_prompt="teste",
        filters={},
        ranking=ranking,
        api_key="fake-key",
    )

    instruction = mock_llm.invoke.call_args[0][0]
    # Extrair o JSON do contexto na instrução
    json_start = instruction.index("{")
    json_str = instruction[json_start:]
    context = json.loads(json_str)
    assert len(context["top_pontos"]) == 5


@patch("core.answer.ChatGroq")
def test_conversao_tipos_numpy(mock_groq_cls):
    """Valores numpy no ranking devem ser serializados sem erro."""
    mock_llm = MagicMock()
    mock_llm.invoke.return_value.content = "resposta"
    mock_groq_cls.return_value = mock_llm

    ranking = [
        {
            "point": "Av. Test, 1",
            "total_target": np.int64(100),
            "total_flow": np.float64(500.0),
            "affinity": np.float64(20.0),
        }
    ]

    # Não deve levantar erro de serialização JSON
    result = generate_final_answer(
        user_prompt="teste numpy",
        filters={"gender": "female"},
        ranking=ranking,
        api_key="fake-key",
    )

    assert result == "resposta"
    # Verificar que invoke foi chamado (serialização não falhou)
    mock_llm.invoke.assert_called_once()
