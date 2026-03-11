import { useEffect, useState } from "react";
import type { ReactNode } from "react";

function Icon({ children }: { children: ReactNode }) {
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
    >
      {children}
    </svg>
  );
}

const SLIDES = [
  "/login/shoppings.jpg",
  "/login/transportes.jpg",
  "/login/elevadores.jpg",
  "/login/ruas.jpg",
  "/login/aeroportos.jpg",
  "/login/rio_galeao.jpg",
];

interface LoginProps {
  onLogin: () => void;
}

export function Login({ onLogin }: LoginProps) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [slideIndex, setSlideIndex] = useState(0);
  const [prevIndex, setPrevIndex] = useState<number | null>(null);

  useEffect(() => {
    const id = setInterval(() => {
      setSlideIndex((current) => {
        setPrevIndex(current);
        return (current + 1) % SLIDES.length;
      });
    }, 4000);
    return () => clearInterval(id);
  }, []);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    onLogin();
  }

  function slideClass(i: number) {
    if (i === slideIndex) return "login-slide login-slide--active";
    if (i === prevIndex) return "login-slide login-slide--prev";
    return "login-slide";
  }

  return (
    <div className="login-shell">
      <div className="login-left">
        <img src="/eletromidia.png" alt="Eletromidia" className="login-logo" />

        <div className="login-card">
          <h2 className="login-title">Entrar na conta</h2>

          <form onSubmit={handleSubmit}>
            <div
              className={`login-field${email ? " login-field--filled" : ""}`}
            >
              <input
                id="login-email"
                type="email"
                placeholder=" "
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                autoComplete="email"
              />
              <label htmlFor="login-email">Seu e-mail</label>
            </div>

            <div
              className={`login-field${password ? " login-field--filled" : ""}`}
            >
              <div className="login-field-inner">
                <input
                  id="login-password"
                  type={showPassword ? "text" : "password"}
                  placeholder=" "
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete="current-password"
                />
                <button
                  type="button"
                  className="login-eye"
                  onClick={() => setShowPassword(!showPassword)}
                  aria-label={showPassword ? "Ocultar senha" : "Mostrar senha"}
                >
                  {showPassword ? (
                    <Icon>
                      <path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94" />
                      <path d="M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19" />
                      <line x1="1" y1="1" x2="23" y2="23" />
                    </Icon>
                  ) : (
                    <Icon>
                      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                      <circle cx="12" cy="12" r="3" />
                    </Icon>
                  )}
                </button>
              </div>
              <label htmlFor="login-password">Senha</label>
            </div>

            <button
              type="submit"
              className="login-submit"
              disabled={!email || !password}
            >
              Entrar
            </button>
          </form>
        </div>
      </div>

      <div className="login-right" aria-hidden>
        {SLIDES.map((src, i) => (
          <img key={src} src={src} alt="" className={slideClass(i)} />
        ))}
      </div>
    </div>
  );
}
