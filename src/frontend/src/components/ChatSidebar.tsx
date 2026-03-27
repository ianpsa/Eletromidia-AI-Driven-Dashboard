import { useCallback, useEffect, useRef, useState } from "react";
import { buildApiUrl } from "../utils/url";

type RankingItem = {
  point: string;
  affinity: number;
  total_target: number;
  total_flow: number;
};

type MessageContent =
  | { type: "text"; text: string }
  | { type: "ranking"; items: RankingItem[] }
  | { type: "dashboard"; url: string };

type Message = {
  id: number;
  sender: "user" | "agent";
  contents: MessageContent[];
  streaming?: boolean;
};

interface ChatSidebarProps {
  open: boolean;
  onClose: () => void;
  onLookerUrl?: (url: string) => void;
}

function parseRankingFromToolResult(content: string): RankingItem[] {
  const items: RankingItem[] = [];
  const lines = content.split("\n");
  for (const line of lines) {
    const match = line.match(
      /^\d+\.\s+(.+?)\s+—\s+Afinidade:\s+([\d.]+)%.*?Público-alvo:\s+([\d.]+).*?Fluxo total:\s+([\d.]+)/,
    );
    if (match) {
      items.push({
        point: match[1],
        affinity: parseFloat(match[2]),
        total_target: parseFloat(match[3]),
        total_flow: parseFloat(match[4]),
      });
    }
  }
  return items;
}

export function ChatSidebar({ open, onClose, onLookerUrl }: ChatSidebarProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [threadId, setThreadId] = useState<string | null>(null);
  const bottomRef = useRef<HTMLDivElement | null>(null);
  const inputRef = useRef<HTMLInputElement | null>(null);

  const messageCount = messages.length;
  useEffect(() => {
    if (messageCount > 0) {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [messageCount]);

  useEffect(() => {
    if (open) {
      inputRef.current?.focus();
    }
  }, [open]);

  const appendToAgentMessage = useCallback(
    (msgId: number, content: MessageContent) => {
      setMessages((prev) =>
        prev.map((m) => {
          if (m.id !== msgId) return m;
          return { ...m, contents: [...m.contents, content] };
        }),
      );
    },
    [],
  );

  const appendTextToken = useCallback((msgId: number, token: string) => {
    setMessages((prev) =>
      prev.map((m) => {
        if (m.id !== msgId) return m;
        const last = m.contents[m.contents.length - 1];
        if (last?.type === "text") {
          return {
            ...m,
            contents: [
              ...m.contents.slice(0, -1),
              { type: "text", text: last.text + token },
            ],
          };
        }
        return {
          ...m,
          contents: [...m.contents, { type: "text", text: token }],
        };
      }),
    );
  }, []);

  const handleSSEEvent = useCallback(
    (msgId: number, event: string, data: Record<string, unknown>) => {
      switch (event) {
        case "metadata":
          setThreadId(data.thread_id as string);
          break;
        case "token":
          appendTextToken(msgId, data.content as string);
          break;
        case "tool_start":
          break;
        case "tool_result": {
          const toolName = data.tool as string;
          const content = data.content as string;
          if (toolName === "analyze_campaign") {
            const ranking = parseRankingFromToolResult(content);
            if (ranking.length > 0) {
              appendToAgentMessage(msgId, { type: "ranking", items: ranking });
            } else {
              appendToAgentMessage(msgId, { type: "text", text: content });
            }
          } else if (toolName === "filter_looker_dashboard") {
            const urlMatch = content.match(/https?:\/\/\S+/);
            if (urlMatch) {
              onLookerUrl?.(urlMatch[0]);
              appendToAgentMessage(msgId, {
                type: "dashboard",
                url: urlMatch[0],
              });
            } else {
              appendToAgentMessage(msgId, { type: "text", text: content });
            }
          } else {
            appendToAgentMessage(msgId, { type: "text", text: content });
          }
          break;
        }
        case "error":
          appendToAgentMessage(msgId, {
            type: "text",
            text: "Erro ao processar solicitação.",
          });
          break;
      }
    },
    [appendTextToken, appendToAgentMessage],
  );

  async function handleSend() {
    if (!input.trim() || loading) return;

    const prompt = input.trim();
    const userMsgId = Date.now();
    const agentMsgId = userMsgId + 1;

    setMessages((prev) => [
      ...prev,
      {
        id: userMsgId,
        sender: "user",
        contents: [{ type: "text", text: prompt }],
      },
      { id: agentMsgId, sender: "agent", contents: [], streaming: true },
    ]);
    setInput("");
    setLoading(true);

    try {
      const res = await fetch(buildApiUrl("/chat"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message: prompt, thread_id: threadId }),
      });

      const reader = res.body!.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop()!;

        let eventType = "";
        for (const line of lines) {
          if (line.startsWith("event: ")) {
            eventType = line.slice(7);
          } else if (line.startsWith("data: ") && eventType) {
            try {
              const data = JSON.parse(line.slice(6));
              handleSSEEvent(agentMsgId, eventType, data);
            } catch {
              /* skip malformed JSON */
            }
            eventType = "";
          }
        }
      }
    } catch {
      setMessages((prev) =>
        prev.map((m) =>
          m.id === agentMsgId
            ? {
                ...m,
                contents: [
                  ...m.contents,
                  { type: "text", text: "Erro ao conectar com o agente." },
                ],
              }
            : m,
        ),
      );
    } finally {
      setMessages((prev) =>
        prev.map((m) => (m.id === agentMsgId ? { ...m, streaming: false } : m)),
      );
      setLoading(false);
    }
  }

  function handleNewChat() {
    setMessages([]);
    setThreadId(null);
  }

  return (
    <aside className={`chat-sidebar ${open ? "chat-sidebar-open" : ""}`}>
      <div className="chat-sidebar-inner">
        <div className="chat-header">
          <div>
            <strong>Agente IA</strong>
          </div>
          <div style={{ display: "flex", gap: 8 }}>
            {messages.length > 0 && (
              <button
                type="button"
                className="chat-close-btn"
                onClick={handleNewChat}
                aria-label="Nova conversa"
                title="Nova conversa"
              >
                +
              </button>
            )}
            <button
              type="button"
              className="chat-close-btn"
              onClick={onClose}
              aria-label="Fechar chat"
            >
              ✕
            </button>
          </div>
        </div>

        <div className="agent-messages">
          {messages.length === 0 && (
            <div className="agent-empty">
              <span className="brand-dot agent-empty-dot" />
              <strong>Como posso ajudar?</strong>
              <p>
                Descreva sua campanha e receba recomendações baseadas em dados
                reais.
              </p>
              <p className="agent-empty-hint">
                Ex: Mulheres 25–34, classe A, perto da Faria Lima
              </p>
            </div>
          )}

          {messages.map((msg) => (
            <div key={msg.id} className={`agent-msg agent-msg-${msg.sender}`}>
              <div
                className={`agent-msg-bubble ${msg.streaming ? "agent-msg-streaming" : ""}`}
              >
                {msg.contents.map((content, i) => {
                  switch (content.type) {
                    case "text":
                      return <p key={i}>{content.text}</p>;
                    case "ranking":
                      return (
                        <div
                          key={i}
                          className="files-table-wrap"
                          style={{ marginTop: 12 }}
                        >
                          <table>
                            <thead>
                              <tr>
                                <th>Ponto</th>
                                <th>Afinidade</th>
                                <th>Público</th>
                                <th>Fluxo</th>
                              </tr>
                            </thead>
                            <tbody>
                              {content.items.map((r, j) => (
                                <tr key={j}>
                                  <td>{r.point}</td>
                                  <td>{r.affinity}%</td>
                                  <td>{r.total_target}</td>
                                  <td>{r.total_flow}</td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      );
                    case "dashboard":
                      return (
                        <div key={i} className="agent-dashboard-embed">
                          <iframe src={content.url} title="Dashboard" />
                        </div>
                      );
                    default:
                      return null;
                  }
                })}

                {msg.streaming && msg.contents.length === 0 && (
                  <div className="agent-typing">
                    <span />
                    <span />
                    <span />
                  </div>
                )}
              </div>
            </div>
          ))}

          <div ref={bottomRef} />
        </div>

        <div className="agent-input-bar">
          <input
            ref={inputRef}
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSend()}
            placeholder="Descreva sua campanha..."
            disabled={loading}
          />
          <button
            type="button"
            onClick={handleSend}
            disabled={!input.trim() || loading}
          >
            Enviar
          </button>
        </div>
      </div>
    </aside>
  );
}
