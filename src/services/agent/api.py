from __future__ import annotations

import json
import os
import uuid

from dotenv import load_dotenv
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import StreamingResponse
from langchain_core.messages import AIMessageChunk, HumanMessage, ToolMessage
from pydantic import BaseModel

from core.agent import get_agent

load_dotenv()

app = FastAPI()

_cors_origins = os.environ.get("CORS_ORIGINS", "http://localhost:5173").split(",")
app.add_middleware(
    CORSMiddleware,
    allow_origins=_cors_origins,
    allow_methods=["POST"],
    allow_headers=["*"],
)


@app.get("/health")
async def health():
    return {"status": "ok"}


class ChatRequest(BaseModel):
    message: str
    thread_id: str | None = None


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
                        payload = json.dumps({"content": event.content})
                        yield f"event: token\ndata: {payload}\n\n"

                if isinstance(event, ToolMessage):
                    payload = json.dumps({"tool": event.name, "content": event.content})
                    yield f"event: tool_result\ndata: {payload}\n\n"
        except Exception:
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
