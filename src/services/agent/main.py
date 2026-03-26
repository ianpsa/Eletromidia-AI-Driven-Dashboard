from __future__ import annotations

import argparse
import asyncio
import os
import sys

from dotenv import load_dotenv

load_dotenv()


async def run(prompt: str, thread_id: str = "cli-session") -> str:
    from langchain_core.messages import HumanMessage

    from core.agent import get_agent

    agent = get_agent()
    config = {"configurable": {"thread_id": thread_id}}
    result = await agent.ainvoke(
        {"messages": [HumanMessage(content=prompt)]},
        config,
    )
    return result["messages"][-1].content


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--prompt", required=True)
    args = parser.parse_args()

    required_keys = {
        "groq": "GROQ_API_KEY",
        "google": "GOOGLE_API_KEY",
    }
    provider = os.getenv("LLM_PROVIDER", "").lower()
    key_name = required_keys.get(provider)
    if key_name and not os.getenv(key_name):
        sys.exit(f"{key_name} not set (LLM_PROVIDER={provider})")

    answer = asyncio.run(run(args.prompt))

    print("\nResposta:\n")
    print(answer)


if __name__ == "__main__":
    main()
