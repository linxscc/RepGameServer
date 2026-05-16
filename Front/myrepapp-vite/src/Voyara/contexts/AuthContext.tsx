import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { setAuthHandlers } from '../api/client';

export interface AuthUser {
  id: number;
  email: string;
  name: string;
  role: string;
  emailVerified: boolean;
  preferredLang: string;
}

interface AuthContextType {
  user: AuthUser | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (token: string, refreshToken: string, user: AuthUser) => void;
  logout: () => void;
  refreshTokens: () => Promise<string | null>;
}

const AuthContext = createContext<AuthContextType>({
  user: null,
  token: null,
  isAuthenticated: false,
  isLoading: true,
  login: () => {},
  logout: () => {},
  refreshTokens: async () => null,
});

function parseJWT(token: string): { exp: number } | null {
  try {
    const payload = token.split('.')[1];
    return JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')));
  } catch {
    return null;
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const doLogout = useCallback(() => {
    localStorage.removeItem('voyara_token');
    localStorage.removeItem('voyara_refresh');
    localStorage.removeItem('voyara_user');
    setToken(null);
    setRefreshToken(null);
    setUser(null);
  }, []);

  const attemptRefresh = useCallback(async (rt: string): Promise<string | null> => {
    try {
      const res = await fetch('/voyara/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refreshToken: rt }),
      });
      if (!res.ok) return null;
      const json = await res.json();
      const data = json.data ?? json;
      localStorage.setItem('voyara_token', data.token);
      localStorage.setItem('voyara_refresh', data.refreshToken);
      setToken(data.token);
      setRefreshToken(data.refreshToken);
      return data.token;
    } catch {
      return null;
    }
  }, []);

  const refreshTokens = useCallback(async (): Promise<string | null> => {
    const rt = refreshToken || localStorage.getItem('voyara_refresh');
    if (!rt) return null;
    return attemptRefresh(rt);
  }, [refreshToken, attemptRefresh]);

  // Wire up 401 interceptor in client.ts
  useEffect(() => {
    setAuthHandlers(refreshTokens, doLogout);
  }, [refreshTokens, doLogout]);

  // Load saved session on mount
  useEffect(() => {
    const savedToken = localStorage.getItem('voyara_token');
    const savedRefresh = localStorage.getItem('voyara_refresh');
    const savedUser = localStorage.getItem('voyara_user');
    if (!savedToken || !savedUser) {
      setIsLoading(false);
      return;
    }

    const decoded = parseJWT(savedToken);
    if (!decoded || decoded.exp * 1000 < Date.now()) {
      if (savedRefresh) {
        attemptRefresh(savedRefresh).then((newToken) => {
          if (!newToken) doLogout();
          setIsLoading(false);
        });
      } else {
        doLogout();
        setIsLoading(false);
      }
      return;
    }

    try {
      setToken(savedToken);
      setRefreshToken(savedRefresh);
      setUser(JSON.parse(savedUser));
    } catch {
      doLogout();
    }
    setIsLoading(false);
  }, []);

  // Periodic expiry check every 5 minutes
  useEffect(() => {
    const interval = setInterval(() => {
      const t = token || localStorage.getItem('voyara_token');
      if (!t) return;
      const decoded = parseJWT(t);
      if (!decoded) return;

      const expMs = decoded.exp * 1000;
      const fiveMin = 5 * 60 * 1000;

      if (expMs < Date.now()) {
        const rt = refreshToken || localStorage.getItem('voyara_refresh');
        if (rt) attemptRefresh(rt).then((newToken) => { if (!newToken) doLogout(); });
        else doLogout();
      } else if (expMs - Date.now() < fiveMin) {
        const rt = refreshToken || localStorage.getItem('voyara_refresh');
        if (rt) attemptRefresh(rt);
      }
    }, 5 * 60 * 1000);

    return () => clearInterval(interval);
  }, [token, refreshToken]);

  const login = useCallback((newToken: string, newRefresh: string, newUser: AuthUser) => {
    localStorage.setItem('voyara_token', newToken);
    if (newRefresh) localStorage.setItem('voyara_refresh', newRefresh);
    localStorage.setItem('voyara_user', JSON.stringify(newUser));
    setToken(newToken);
    setRefreshToken(newRefresh);
    setUser(newUser);
  }, []);

  return (
    <AuthContext.Provider value={{
      user, token, isAuthenticated: !!token && !!user, isLoading,
      login, logout: doLogout, refreshTokens,
    }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
