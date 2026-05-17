import { createContext, useContext, useState, useEffect, useCallback, useRef, type ReactNode } from "react";
import { api } from "@/lib/api";

interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  loading: boolean;
  setupRequired: boolean | null; // null = still checking
  login: (email: string, password: string) => Promise<void>;
  setup: (email: string, password: string, name: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(() =>
    localStorage.getItem("throtl_token")
  );
  const [setupRequired, setSetupRequired] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(true);
  const tokenRef = useRef(token);
  tokenRef.current = token;

  // Check setup status on mount
  useEffect(() => {
    let cancelled = false;

    async function init() {
      try {
        const res = await api.checkSetup();
        if (cancelled) return;

        setSetupRequired(res.setup_required);

        // If admin exists and we have a stored token, verify it
        if (!res.setup_required && tokenRef.current) {
          try {
            const u = await api.getMe();
            if (!cancelled) setUser(u);
          } catch {
            // Token invalid — clear it
            localStorage.removeItem("throtl_token");
            if (!cancelled) setToken(null);
          }
        }
      } catch {
        // API unreachable — assume setup needed (safer default)
        if (!cancelled) setSetupRequired(true);
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    init();
    return () => { cancelled = true; };
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await api.login({ email, password });
    localStorage.setItem("throtl_token", res.token);
    setToken(res.token);
    setUser(res.user);
  }, []);

  const setup = useCallback(async (email: string, password: string, name: string) => {
    const res = await api.setup({ email, password, name });
    localStorage.setItem("throtl_token", res.token);
    setToken(res.token);
    setUser(res.user);
    setSetupRequired(false);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem("throtl_token");
    setToken(null);
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, token, loading, setupRequired, login, setup, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
