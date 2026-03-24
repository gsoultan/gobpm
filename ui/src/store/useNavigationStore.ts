/**
 * useNavigationStore — manages the user's current project/organization context.
 *
 * FE-ARCH-2: Separated from the monolithic useAppStore to satisfy SRP.
 * Persisted to localStorage under the key 'gobpm-nav-storage'.
 */
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

interface NavigationState {
  currentProjectId: string | null;
  setCurrentProjectId: (id: string | null) => void;
  currentOrganizationId: string | null;
  setCurrentOrganizationId: (id: string | null) => void;
}

export const useNavigationStore = create<NavigationState>()(
  persist(
    (set) => ({
      currentProjectId: null,
      setCurrentProjectId: (id) => set({ currentProjectId: id }),
      currentOrganizationId: null,
      setCurrentOrganizationId: (id) => set({ currentOrganizationId: id }),
    }),
    {
      name: 'gobpm-nav-storage',
      storage: createJSONStorage(() => localStorage),
    },
  ),
);

