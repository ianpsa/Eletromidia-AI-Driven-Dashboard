from __future__ import annotations

import json
import logging
import os
import uuid

from dotenv import load_dotenv
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import StreamingResponse
from langchain_core.messages import AIMessageChunk, HumanMessage, ToolMessage
from pydantic import BaseModel, Field

from core.agent import get_agent

load_dotenv()

PORT = int(os.environ.get("PORT", "8001"))

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(name)s %(levelname)s %(message)s",
)
logger = logging.getLogger(__name__)

app = FastAPI()

_cors_origins = os.environ.get("CORS_ORIGINS", "http://localhost:5173").split(",")
app.add_middleware(
    CORSMiddleware,
    allow_origins=_cors_origins,
    allow_methods=["POST"],
    allow_headers=["Content-Type"],
)


@app.get("/health")
async def health():
    return {"status": "ok"}


class ChatRequest(BaseModel):
    message: str = Field(..., min_length=1, max_length=10_000)
    thread_id: str | None = Field(
        default=None, max_length=100, pattern=r"^[a-zA-Z0-9_\-]+$"
    )


@app.post("/chat")
async def chat(request: ChatRequest):
    thread_id = request.thread_id or str(uuid.uuid4())
    agent = get_agent()

    async def event_stream():
        yield f"event: metadata\ndata: {json.dumps({'thread_id': thread_id})}\n\n"

        config = {"configurable": {"thread_id": thread_id}}
        input_msg = {"messages": [HumanMessage(content=request.message)]}

        try:
            async for event, _metadata in agent.astream(
                input_msg, config, stream_mode="messages"
            ):
                if isinstance(event, AIMessageChunk):
                    if event.tool_call_chunks:
                        for chunk in event.tool_call_chunks:
                            if chunk.get("name"):
                                payload = json.dumps({"tool": chunk["name"]})
                                yield f"event: tool_start\ndata: {payload}\n\n"
                    elif event.content:
                        text = event.content
                        if isinstance(text, list):
                            text = "".join(
                                block.get("text", "")
                                for block in text
                                if isinstance(block, dict)
                            )
                        if text:
                            payload = json.dumps({"content": text})
                            yield f"event: token\ndata: {payload}\n\n"

                if isinstance(event, ToolMessage):
                    if event.name == "filter_looker_dashboard":
                        try:
                            parsed = json.loads(event.content)
                            dashboard_payload = json.dumps(
                                {
                                    "url": parsed["url"],
                                    "filters": parsed["filters_applied"],
                                }
                            )
                            yield (
                                f"event: dashboard_update\n"
                                f"data: {dashboard_payload}\n\n"
                            )

                            chat_payload = json.dumps(
                                {
                                    "tool": event.name,
                                    "content": parsed["summary"],
                                }
                            )
                            yield f"event: tool_result\ndata: {chat_payload}\n\n"
                        except (json.JSONDecodeError, KeyError):
                            payload = json.dumps(
                                {"tool": event.name, "content": event.content}
                            )
                            yield f"event: tool_result\ndata: {payload}\n\n"
                    else:
                        payload = json.dumps(
                            {"tool": event.name, "content": event.content}
                        )
                        yield f"event: tool_result\ndata: {payload}\n\n"
        except Exception:
            logger.exception("SSE stream error for thread_id=%s", thread_id)
            payload = json.dumps({"error": "Erro interno. Tente novamente."})
            yield f"event: error\ndata: {payload}\n\n"

        yield "event: done\ndata: {}\n\n"

    return StreamingResponse(
        event_stream(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "X-Accel-Buffering": "no",
        },
    )


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("api:app", host="0.0.0.0", port=PORT, reload=True)
