import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react';
import { getMe, googleLogin, setUnauthorizedHandler } from '../api/client';
import { clearStoredToken, getStoredToken, setStoredToken } from './tokenStorage';
import type { User } from '../types/api';

type AuthState =
  | { status: 'loading' }
  | { status: 'signedOut' }
  | { status: 'signedIn'; user: User };

type AuthContextValue = {
  state: AuthState;
  signInWithGoogleIdToken: (idToken: string) => Promise<void>;
  signOut: () => Promise<void>;
  handleUnauthorized: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ status: 'loading' });

  useEffect(() => {
    (async () => {
      const token = await getStoredToken();
      if (!token) {
        setState({ status: 'signedOut' });
        return;
      }
      try {
        const user = await getMe();
        setState({ status: 'signedIn', user });
      } catch {
        // Stored token is missing, expired, or the backend rejected it —
        // either way there's no valid session to resume.
        await clearStoredToken();
        setState({ status: 'signedOut' });
      }
    })();
  }, []);

  const signInWithGoogleIdToken = useCallback(async (idToken: string) => {
    const { token, user } = await googleLogin(idToken);
    await setStoredToken(token);
    setState({ status: 'signedIn', user });
  }, []);

  const signOut = useCallback(async () => {
    await clearStoredToken();
    setState({ status: 'signedOut' });
  }, []);

  const handleUnauthorized = useCallback(async () => {
    await clearStoredToken();
    setState({ status: 'signedOut' });
  }, []);

  useEffect(() => {
    setUnauthorizedHandler(() => {
      handleUnauthorized();
    });
  }, [handleUnauthorized]);

  return (
    <AuthContext.Provider value={{ state, signInWithGoogleIdToken, signOut, handleUnauthorized }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
