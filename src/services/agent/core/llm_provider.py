from __future__ import annotations

import os
from typing import Protocol

from langchain_core.language_models import BaseChatModel


class LLMProvider(Protocol):
    def build(self) -> BaseChatModel: ...


class GroqProvider:
    def build(self) -> BaseChatModel:
        from langchain_groq import ChatGroq

        return ChatGroq(
            api_key=os.environ["GROQ_API_KEY"],
            model_name=os.environ.get("LLM_MODEL", "llama-3.3-70b-versatile"),
            temperature=float(os.environ.get("LLM_TEMPERATURE", "0.3")),
        )


class GoogleProvider:
    def build(self) -> BaseChatModel:
        from langchain_google_genai import ChatGoogleGenerativeAI

        return ChatGoogleGenerativeAI(
            google_api_key=os.environ["GOOGLE_API_KEY"],
            model=os.environ.get("LLM_MODEL", "gemini-2.0-flash"),
            temperature=float(os.environ.get("LLM_TEMPERATURE", "0.3")),
        )


_PROVIDERS: dict[str, type[LLMProvider]] = {
    "groq": GroqProvider,
    "google": GoogleProvider,
}

_cached_provider: LLMProvider | None = None


def get_llm_provider() -> LLMProvider:
    global _cached_provider
    if _cached_provider is not None:
        return _cached_provider

    name = os.environ.get("LLM_PROVIDER", "groq").lower()
    cls = _PROVIDERS.get(name)
    if cls is None:
        raise ValueError(
            f"Unknown LLM_PROVIDER={name!r}. Choose from: {', '.join(_PROVIDERS)}"
        )
    _cached_provider = cls()
    return _cached_provider
