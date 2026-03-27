import json
from unittest.mock import MagicMock, patch

import pytest

from core.llm import parse_prompt


@patch("core.llm.ChatGroq")
def test_parse_prompt_json_valido(mock_groq_cls):
    """Resposta JSON válida da LLM é parseada corretamente."""
    mock_llm = MagicMock()
    mock_llm.invoke.return_value.content = '{"gender": "female", "age_min": 18, "age_max": 25}'
    mock_groq_cls.return_value = mock_llm

    result = parse_prompt("campanha para mulheres jovens", "fake-key")

    assert result == {"gender": "female", "age_min": 18, "age_max": 25}


@patch("core.llm.ChatGroq")
def test_parse_prompt_resposta_vazia_levanta_system_exit(mock_groq_cls):
    """Resposta vazia da LLM deve levantar SystemExit."""
    mock_llm = MagicMock()
    mock_llm.invoke.return_value.content = ""
    mock_groq_cls.return_value = mock_llm

    with pytest.raises(SystemExit, match="Resposta vazia do GROQ"):
        parse_prompt("qualquer prompt", "fake-key")


@patch("core.llm.ChatGroq")
def test_parse_prompt_json_invalido_levanta_erro(mock_groq_cls):
    """JSON inválido na resposta da LLM deve levantar JSONDecodeError."""
    mock_llm = MagicMock()
    mock_llm.invoke.return_value.content = "isto não é json"
    mock_groq_cls.return_value = mock_llm

    with pytest.raises(json.JSONDecodeError):
        parse_prompt("qualquer prompt", "fake-key")


@patch("core.llm.ChatGroq")
def test_parse_prompt_instrucao_contem_prompt_usuario(mock_groq_cls):
    """A instrução enviada à LLM deve conter o prompt original do usuário."""
    mock_llm = MagicMock()
    mock_llm.invoke.return_value.content = '{"gender": "male"}'
    mock_groq_cls.return_value = mock_llm

    user_prompt = "campanha para homens classe A em Curitiba"
    parse_prompt(user_prompt, "fake-key")

    call_args = mock_llm.invoke.call_args
    instruction = call_args[0][0]
    assert user_prompt in instruction


@patch("core.llm.ChatGroq")
def test_parse_prompt_resposta_none_levanta_system_exit(mock_groq_cls):
    """Resposta None da LLM deve levantar SystemExit."""
    mock_llm = MagicMock()
    mock_llm.invoke.return_value.content = None
    mock_groq_cls.return_value = mock_llm

    with pytest.raises(SystemExit, match="Resposta vazia do GROQ"):
        parse_prompt("qualquer prompt", "fake-key")
