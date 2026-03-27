from __future__ import annotations

import os
import threading
from typing import Protocol

from langchain_core.language_models import BaseChatModel


class LLMProvider(Protocol):
    def build(self) -> BaseChatModel: ...


class GroqProvider:
    def build(self) -> BaseChatModel:
        from langchain_groq import ChatGroq

        return ChatGroq(
            api_key=os.environ["GROQ_API_KEY"],
            model_name=os.environ.get("LLM_MODEL", "llama3-groq-70b-8192-tool-use-preview"),
            temperature=float(os.environ.get("LLM_TEMPERATURE", "0.3")),
            streaming=False,
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
_provider_lock = threading.Lock()


def get_llm_provider() -> LLMProvider:
    global _cached_provider
    if _cached_provider is not None:
        return _cached_provider
    with _provider_lock:
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
