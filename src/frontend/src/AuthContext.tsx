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

interface AuthState {
  user: User | null;
  idToken: string | null;
  loading: boolean;
  error: string | null;
}

interface AuthContextValue extends AuthState {
  signInWithGoogle: () => Promise<void>;
  signOut: () => Promise<void>;
  /** Always returns a fresh token (auto-refreshed by Firebase ~every 55 min) */
  getToken: () => Promise<string>;
}

// ─── Context ─────────────────────────────────────────────────────────────────

const AuthContext = createContext<AuthContextValue | null>(null);

// ─── Provider ────────────────────────────────────────────────────────────────

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    user: null,
    idToken: null,
    loading: true,
    error: null,
  });

  // Keep a ref so getToken() always sees the latest user without a stale closure
  const userRef = useRef<User | null>(null);

  // Subscribe to Firebase token changes (covers initial load + auto-refresh)
  useEffect(() => {
    const unsubscribe = onIdTokenChanged(auth, async (firebaseUser) => {
      userRef.current = firebaseUser;

      if (firebaseUser) {
        const token = await firebaseUser.getIdToken();
        setState({ user: firebaseUser, idToken: token, loading: false, error: null });
      } else {
        setState({ user: null, idToken: null, loading: false, error: null });
      }
    });

    return unsubscribe;
  }, []);

  const signInWithGoogle = useCallback(async () => {
    setState((s) => ({ ...s, loading: true, error: null }));
    try {
      await signInWithPopup(auth, googleProvider);
      // onIdTokenChanged fires automatically after sign-in
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Falha ao entrar com Google";
      setState((s) => ({ ...s, loading: false, error: msg }));
    }
  }, []);

  const signOut = useCallback(async () => {
    await firebaseSignOut(auth);
  }, []);

  /** Returns a valid (possibly refreshed) token — safe to call before every request */
  const getToken = useCallback(async (): Promise<string> => {
    const u = userRef.current;
    if (!u) throw new Error("Usuário não autenticado");
    return u.getIdToken(); // Firebase only refreshes if token is < 5 min from expiry
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