import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import {
  type User,
  onIdTokenChanged,
  signInWithPopup,
  signOut as firebaseSignOut,
} from "firebase/auth";
import { auth, googleProvider } from "./firebase";

// ─── Types ──────────────────────────────────────────────────────────────────

const AUTH_BASE = ((import.meta.env.VITE_BASE_URL as string) || "").replace(/\/$/, "");

type AuthStatus = "idle" | "authorized" | "unauthorized";

interface AuthState {
  user: User | null;
  idToken: string | null;
  loading: boolean;
  error: string | null;
  status: AuthStatus;
}

interface AuthContextValue extends AuthState {
  signInWithGoogle: () => Promise<void>;
  signOut: () => Promise<void>;
  /** Always returns a fresh token (auto-refreshed by Firebase ~every 55 min) */
  getToken: () => Promise<string>;
}

// ─── Context ─────────────────────────────────────────────────────────────────

const AuthContext = createContext<AuthContextValue | null>(null);

// ─── Helpers ─────────────────────────────────────────────────────────────────

async function callValidate(token: string): Promise<AuthStatus> {
  try {
    const res = await fetch(`${AUTH_BASE}/auth/validate`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) return "authorized";
    return "unauthorized";
  } catch {
    return "unauthorized";
  }
}

// ─── Provider ────────────────────────────────────────────────────────────────

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    user: null,
    idToken: null,
    loading: true,
    error: null,
    status: "idle",
  });

  const userRef = useRef<User | null>(null);

  // Resolve persisted session on page load
  useEffect(() => {
    const unsubscribe = onIdTokenChanged(auth, async (firebaseUser) => {
      userRef.current = firebaseUser;

      if (firebaseUser) {
        const token = await firebaseUser.getIdToken();
        const status = await callValidate(token);

        if (status === "authorized") {
          setState({ user: firebaseUser, idToken: token, loading: false, error: null, status });
        } else {
          // Token válido no Firebase mas sem role no IAM — desloga automaticamente
          await firebaseSignOut(auth);
          setState({
            user: null,
            idToken: null,
            loading: false,
            error: "Acesso não autorizado. Seu email não tem permissão para acessar este sistema.",
            status: "unauthorized",
          });
        }
      } else {
        setState({ user: null, idToken: null, loading: false, error: null, status: "idle" });
      }
    });

    return unsubscribe;
  }, []);

  const signInWithGoogle = useCallback(async () => {
    setState((s) => ({ ...s, loading: true, error: null }));
    try {
      const result = await signInWithPopup(auth, googleProvider);
      const token = await result.user.getIdToken();
      const status = await callValidate(token);

      if (status === "authorized") {
        // onIdTokenChanged vai disparar e setar o estado completo
        return;
      }

      // Sem role — desloga e mostra erro
      await firebaseSignOut(auth);
      setState({
        user: null,
        idToken: null,
        loading: false,
        error: "Acesso não autorizado. Seu email não tem permissão para acessar este sistema.",
        status: "unauthorized",
      });
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Falha ao entrar com Google";
      setState((s) => ({ ...s, loading: false, error: msg, status: "idle" }));
    }
  }, []);

  const signOut = useCallback(async () => {
    await firebaseSignOut(auth);
    setState({ user: null, idToken: null, loading: false, error: null, status: "idle" });
  }, []);

  const getToken = useCallback(async (): Promise<string> => {
    const u = userRef.current;
    if (!u) throw new Error("Usuário não autenticado");
    return u.getIdToken();
  }, []);

  return (
    <AuthContext.Provider value={{ ...state, signInWithGoogle, signOut, getToken }}>
      {children}
    </AuthContext.Provider>
  );
}

// ─── Hook ────────────────────────────────────────────────────────────────────

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside <AuthProvider>");
  return ctx;
}