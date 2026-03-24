/**
 * useAuthStore — manages authentication state (user, JWT token).
 *
 * FE-ARCH-2: Separated from the monolithic useAppStore to satisfy SRP.
 * Persisted to localStorage under the key 'gobpm-auth-storage'.
 */
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

export interface AuthUser {
  id: string;
  name: string;
  username: string;
  role: string;
  organizations?: Array<{ id: string; name: string }>;
  projects?: Array<{ id: string; name: string }>;
}

interface AuthState {
  user: AuthUser | null;
  token: string | null;
  setAuth: (user: AuthUser, token: string) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      setAuth: (user, token) => set({ user, token }),
      clearAuth: () => set({ user: null, token: null }),
    }),
    {
      name: 'gobpm-auth-storage',
      storage: createJSONStorage(() => localStorage),
    },
  ),
);

