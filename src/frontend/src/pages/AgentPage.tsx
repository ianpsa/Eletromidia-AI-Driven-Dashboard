import { useState, useRef, useEffect } from "react";
import { buildApiUrl } from "../utils/url";

type RankingItem = {
  point: string;
  affinity: number;
  total_target: number;
  total_flow: number;
};

type Message = {
  id: number;
  sender: "user" | "agent";
  text: string;
  ranking?: RankingItem[];
};

export function AgentPage() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const bottomRef = useRef<HTMLDivElement | null>(null);
  const inputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  async function handleSend() {
    if (!input.trim() || loading) return;

    const prompt = input.trim();
    setMessages((m) => [
      ...m,
      { id: Date.now(), sender: "user", text: prompt },
    ]);
    setInput("");
    setLoading(true);

    try {
      const res = await fetch(buildApiUrl("/analyze"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt, limit: 5 }),
      });
      const data = await res.json();

      setMessages((m) => [
        ...m,
        {
          id: Date.now() + 1,
          sender: "agent",
          text: data.analysis || data.message || "Análise concluída.",
          ranking: data.top_points,
        },
      ]);
    } catch {
      setMessages((m) => [
        ...m,
        {
          id: Date.now() + 2,
          sender: "agent",
          text: "Erro ao processar solicitação.",
        },
      ]);
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <header className="topbar">
        <div>
          <h1>Agente IA</h1>
          <p>Análise estratégica de mídia OOH</p>
        </div>
      </header>

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
              Ex: Mulheres 25–34, classe A, São Paulo
            </p>
          </div>
        )}

        {messages.map((msg) => (
          <div key={msg.id} className={`agent-msg agent-msg-${msg.sender}`}>
            <div className="agent-msg-bubble">
              <p>{msg.text}</p>

              {msg.ranking && msg.ranking.length > 0 && (
                <div className="files-table-wrap" style={{ marginTop: 12 }}>
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
                      {msg.ranking.map((r, i) => (
                        <tr key={i}>
                          <td>{r.point}</td>
                          <td>{r.affinity}%</td>
                          <td>{r.total_target}</td>
                          <td>{r.total_flow}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        ))}

        {loading && (
          <div className="agent-msg agent-msg-agent">
            <div className="agent-msg-bubble">
              <div className="agent-typing">
                <span />
                <span />
                <span />
              </div>
            </div>
          </div>
        )}

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
        <button onClick={handleSend} disabled={!input.trim() || loading}>
          Enviar
        </button>
      </div>
    </>
  );
}
