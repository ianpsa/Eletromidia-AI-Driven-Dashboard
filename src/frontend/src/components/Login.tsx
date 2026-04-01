import { useEffect, useState } from "react";
import { useAuth } from "../AuthContext";

const SLIDES = [
  "/login/shoppings.jpg",
  "/login/transportes.jpg",
  "/login/elevadores.jpg",
  "/login/ruas.jpg",
  "/login/aeroportos.jpg",
  "/login/rio_galeao.jpg",
];

export function Login() {
  const { signInWithGoogle, loading, error } = useAuth();
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
          <p className="login-subtitle">
            Use sua conta Google corporativa para acessar.
          </p>

          {error && (
            <div className="login-error" role="alert">
              {error}
            </div>
          )}

          <button
            type="button"
            className="login-google-btn"
            onClick={signInWithGoogle}
            disabled={loading}
          >
            <GoogleIcon />
            {loading ? "Entrando…" : "Entrar com Google"}
          </button>
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

function GoogleIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 48 48" aria-hidden>
      <path
        fill="#FFC107"
        d="M43.6 20.1H42V20H24v8h11.3C33.6 32.7 29.2 36 24 36c-6.6 0-12-5.4-12-12s5.4-12 12-12c3.1 0 5.8 1.1 8 2.9l5.7-5.7C34.5 6.5 29.6 4 24 4 12.9 4 4 12.9 4 24s8.9 20 20 20 20-8.9 20-20c0-1.3-.1-2.7-.4-3.9z"
      />
      <path
        fill="#FF3D00"
        d="M6.3 14.7l6.6 4.8C14.5 15.1 18.9 12 24 12c3.1 0 5.8 1.1 8 2.9l5.7-5.7C34.5 6.5 29.6 4 24 4 16.3 4 9.7 8.3 6.3 14.7z"
      />
      <path
        fill="#4CAF50"
        d="M24 44c5.2 0 9.9-2 13.4-5.2l-6.2-5.2C29.3 35.5 26.8 36 24 36c-5.2 0-9.6-3.3-11.3-8H6.3C9.7 35.6 16.3 44 24 44z"
      />
      <path
        fill="#1976D2"
        d="M43.6 20.1H42V20H24v8h11.3c-.9 2.4-2.5 4.5-4.5 6l6.2 5.2C40.2 35.9 44 30.4 44 24c0-1.3-.1-2.7-.4-3.9z"
      />
    </svg>
  );
}
