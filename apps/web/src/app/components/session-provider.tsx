'use client';

import { createContext, useContext, useEffect, useMemo, useState } from 'react';

const SESSION_STORAGE_KEY = 'urbanmemory-session';

export type SessionMode = 'public' | 'admin';

export type AdminSession = {
  adminId: number;
  email: string;
  role: string;
  token: string;
};

export type SessionState = {
  mode: SessionMode;
  admin: AdminSession | null;
  isReady: boolean;
};

type SessionContextValue = {
  session: SessionState;
  setPublicSession: () => void;
  setAdminSession: (admin: AdminSession) => void;
  clearSession: () => void;
};

const SessionContext = createContext<SessionContextValue | null>(null);

const DEFAULT_SESSION: SessionState = {
  mode: 'public',
  admin: null,
  isReady: false,
};

export function SessionProvider({ children }: { children: React.ReactNode }) {
  const [session, setSession] = useState<SessionState>(DEFAULT_SESSION);

  useEffect(() => {
    try {
      const stored = window.localStorage.getItem(SESSION_STORAGE_KEY);
      if (!stored) {
        setSession((current) => ({ ...current, isReady: true }));
        return;
      }

      const parsed = JSON.parse(stored) as Partial<SessionState>;
      if (parsed?.mode === 'admin' && parsed.admin?.token) {
        setSession({
          mode: 'admin',
          admin: parsed.admin,
          isReady: true,
        });
        return;
      }

      setSession({
        mode: 'public',
        admin: null,
        isReady: true,
      });
    } catch {
      setSession({
        mode: 'public',
        admin: null,
        isReady: true,
      });
    }
  }, []);

  useEffect(() => {
    if (!session.isReady) {
      return;
    }

    window.localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session));
  }, [session]);

  const value = useMemo<SessionContextValue>(() => {
    return {
      session,
      setPublicSession: () => {
        setSession({ mode: 'public', admin: null, isReady: true });
      },
      setAdminSession: (admin) => {
        setSession({ mode: 'admin', admin, isReady: true });
      },
      clearSession: () => {
        setSession({ mode: 'public', admin: null, isReady: true });
      },
    };
  }, [session]);

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession() {
  const context = useContext(SessionContext);
  if (!context) {
    throw new Error('useSession must be used within SessionProvider');
  }

  return context;
}