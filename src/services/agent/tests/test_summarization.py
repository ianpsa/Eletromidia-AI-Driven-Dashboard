"""Tests for pure functions in core.summarization."""

from __future__ import annotations

from langchain_core.messages import HumanMessage

from core.summarization import _estimate_tokens, _get_context_window


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


# ── _get_context_window ───────────────────────────────────────────────


class TestGetContextWindow:
    def test_groq_model(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "llama-3.3-70b-versatile")
        assert _get_context_window() == 128_000

    def test_gemini_model(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "gemini-2.0-flash")
        assert _get_context_window() == 1_048_576

    def test_unknown_model_returns_default(self, monkeypatch):
        monkeypatch.setenv("LLM_MODEL", "modelo-desconhecido")
        assert _get_context_window() == 128_000
