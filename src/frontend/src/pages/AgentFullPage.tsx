import { useState, useRef, useEffect } from "react";
import { TopBar } from "../components/TopBar";
import { buildApiUrl } from "../utils/url";

type Message = {
  id: number;
  sender: "user" | "agent";
  text: string;
  details?: {
    ranking?: Array<{
      point: string;
      affinity: number;
      total_target: number;
      total_flow: number;
    }>;
  };
};

export default function AgentFullPage() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 0,
      sender: "agent",
      text: "Descreva sua campanha e eu trarei recomendações estratégicas com base em dados.",
    },
  ]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);

  const bottomRef = useRef<HTMLDivElement | null>(null);

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
          text:
            data.analysis ||
            data.message ||
            "Não foi possível gerar análise.",
          details: data.top_points
            ? { ranking: data.top_points }
            : undefined,
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
      <TopBar
        bucketName="Agente IA"
        query=""
        onQueryChange={() => {}}
        onRefresh={() => {}}
      />

      <section className="agent-section">
        <header className="agent-section-header">
          <h2>Análise estratégica de mídia OOH</h2>
          <p>
            Utilize linguagem natural para descrever seu público e receba
            recomendações baseadas em dados reais.
          </p>
        </header>

        <div className="agent-card">
          <div className="agent-messages">
            {messages.map((msg) => (
              <div
                key={msg.id}
                className={`agent-row ${msg.sender}`}
              >
                <div className="agent-bubble">
                  {msg.text}

                  {msg.details?.ranking && (
                    <div className="agent-ranking">
                      {msg.details.ranking.map((p, i) => (
                        <div key={i}>
                          <strong>{p.point}</strong>
                          <span>
                            Afinidade: {p.affinity}% • Público:{" "}
                            {p.total_target} • Fluxo: {p.total_flow}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            ))}

            {loading && (
              <div className="agent-row agent">
                <div className="agent-bubble typing">
                  Analisando...
                </div>
              </div>
            )}

            <div ref={bottomRef} />
          </div>

          <div className="agent-input">
            <input
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Ex: Mulheres 25-34, classe A, São Paulo"
              onKeyDown={(e) => e.key === "Enter" && handleSend()}
              disabled={loading}
            />
            <button
              onClick={handleSend}
              disabled={!input.trim() || loading}
            >
              Enviar
            </button>
          </div>
        </div>
      </section>
    </>
  );
}