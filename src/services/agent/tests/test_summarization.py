"""Tests for pure functions in core.summarization."""

from __future__ import annotations

from unittest.mock import MagicMock, patch

from langchain_core.messages import AIMessage, HumanMessage, SystemMessage

from core.summarization import _estimate_tokens, _get_context_window, maybe_summarize

# ── _estimate_tokens ──────────────────────────────────────────────────


class TestEstimateTokens:
    def test_simple_string(self):
        msg = HumanMessage(content="abcd" * 10)  # 40 chars → 10 tokens
        assert _estimate_tokens([msg]) == 10

    def test_content_list_of_dicts(self):
        msg = HumanMessage(content=[{"text": "a" * 80}, {"text": "b" * 40}])
        # 80/4 + 40/4 = 30
        assert _estimate_tokens([msg]) == 30

    def test_empty_list(self):
        assert _estimate_tokens([]) == 0

    def test_empty_string_content(self):
        msg = HumanMessage(content="")
        assert _estimate_tokens([msg]) == 0

    def test_content_list_of_strings(self):
        msg = HumanMessage(content=["a" * 40, "b" * 40])
        # 40/4 + 40/4 = 20
        assert _estimate_tokens([msg]) == 20

    def test_multiple_messages_summed(self):
        msgs = [
            HumanMessage(content="a" * 40),   # 10 tokens
            AIMessage(content="b" * 80),       # 20 tokens
        ]
        assert _estimate_tokens(msgs) == 30

    def test_dict_without_text_key_ignored(self):
        msg = HumanMessage(content=[{"image_url": "http://example.com/img.png"}])
        assert _estimate_tokens([msg]) == 0

    def test_system_message_counted(self):
        msg = SystemMessage(content="x" * 100)
        assert _estimate_tokens([msg]) == 25


# ── _get_context_window ───────────────────────────────────────────────


class TestGetContextWindow:
    def test_groq_model(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "llama-3.3-70b-versatile")
        assert _get_context_window() == 128_000

    def test_groq_model_v2(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "llama-3.1-70b-versatile")
        assert _get_context_window() == 128_000

    def test_gemini_model(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "gemini-2.0-flash")
        assert _get_context_window() == 1_000_000

    def test_anthropic_haiku(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "claude-haiku-4-5")
        assert _get_context_window() == 200_000

    def test_anthropic_sonnet(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "claude-sonnet-4-5")
        assert _get_context_window() == 200_000

    def test_unknown_model_returns_default(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "modelo-desconhecido")
        assert _get_context_window() == 128_000

    def test_empty_model_returns_default(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "")
        assert _get_context_window() == 128_000


# ── maybe_summarize ───────────────────────────────────────────────────


def _make_messages(n_human: int, chars_each: int = 100):
    """Build a list of alternating Human/AI messages."""
    msgs = []
    for i in range(n_human):
        msgs.append(HumanMessage(content="h" * chars_each, id=f"h{i}"))
        msgs.append(AIMessage(content="a" * chars_each, id=f"a{i}"))
    return msgs


class TestMaybeSummarize:
    def test_below_threshold_returns_empty(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "llama-3.3-70b-versatile")
        # 128_000 * 0.75 = 96_000 tokens threshold; small messages stay below
        msgs = _make_messages(n_human=3, chars_each=100)
        result = maybe_summarize({"messages": msgs})
        assert result == {"messages": []}

    def test_above_threshold_calls_llm(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "llama-3.3-70b-versatile")
        # Each message has 96_001 * 4 chars to exceed the 96_000 token threshold
        big_content = "x" * (96_001 * 4)
        msgs = [
            HumanMessage(content=big_content, id="h0"),
            AIMessage(content="response", id="a0"),
            HumanMessage(content="follow-up", id="h1"),
        ]
        mock_llm = MagicMock()
        mock_llm.invoke.return_value = MagicMock(content="Summary text")

        import core.summarization as summ_mod

        original = summ_mod._cached_summarization_llm
        summ_mod._cached_summarization_llm = mock_llm
        try:
            result = maybe_summarize({"messages": msgs})
        finally:
            summ_mod._cached_summarization_llm = original

        mock_llm.invoke.assert_called_once()
        # Result should contain RemoveMessage entries and a new SystemMessage summary
        assert len(result["messages"]) > 0

    def test_empty_messages_returns_empty(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "llama-3.3-70b-versatile")
        result = maybe_summarize({"messages": []})
        assert result == {"messages": []}
