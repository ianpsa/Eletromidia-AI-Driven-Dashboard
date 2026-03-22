from __future__ import annotations

from typing import Literal

from langchain_core.messages import SystemMessage
from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import START, MessagesState, StateGraph
from langgraph.prebuilt import ToolNode

from core.llm_provider import get_llm_provider
from core.prompt import SYSTEM_PROMPT
from core.summarization import maybe_summarize
from core.tools.bigquery import query_bigquery
from core.tools.campaign import analyze_campaign
from core.tools.geocoding import geocode_location
from core.tools.looker import filter_looker_dashboard
from core.tools.metadata import get_available_filters

TOOLS = [
    analyze_campaign,
    geocode_location,
    get_available_filters,
    query_bigquery,
    filter_looker_dashboard,
]


_cached_llm = None


def _get_llm():
    global _cached_llm
    if _cached_llm is None:
        _cached_llm = get_llm_provider().build().bind_tools(TOOLS)
    return _cached_llm


def llm_call(state: MessagesState) -> dict:
    """Invoke the LLM with the current message history."""
    llm = _get_llm()
    messages = [SystemMessage(content=SYSTEM_PROMPT)] + list(state["messages"])
    response = llm.invoke(messages)
    return {"messages": [response]}


def should_continue(state: MessagesState) -> Literal["tool_node", "__end__"]:
    """Route to tools if the LLM made tool calls, otherwise end."""
    last = state["messages"][-1]
    if hasattr(last, "tool_calls") and last.tool_calls:
        return "tool_node"
    return "__end__"


def build_graph() -> StateGraph:
    """Construct the agent graph (not compiled)."""
    tool_node = ToolNode(tools=TOOLS)

    graph = StateGraph(MessagesState)
    graph.add_node("maybe_summarize", maybe_summarize)
    graph.add_node("llm_call", llm_call)
    graph.add_node("tool_node", tool_node)

    graph.add_edge(START, "maybe_summarize")
    graph.add_edge("maybe_summarize", "llm_call")
    graph.add_conditional_edges("llm_call", should_continue)
    graph.add_edge("tool_node", "llm_call")

    return graph


_memory = MemorySaver()
_cached_agent = None


def get_agent():
    """Return the compiled agent with memory checkpointer."""
    global _cached_agent
    if _cached_agent is None:
        _cached_agent = build_graph().compile(checkpointer=_memory)
    return _cached_agent
