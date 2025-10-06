import { create } from "zustand";
import { User as FirebaseUser } from "firebase/auth";
import {
  signInWithGoogle,
  signInWithGithub,
  signOut as firebaseSignOut,
  onAuthStateChanged,
  getIdToken,
} from "@/lib/auth";
import { apiClient, BackendUser } from "@/lib/api";

interface AuthState {
  // State
  firebaseUser: FirebaseUser | null;
  backendUser: BackendUser | null;
  firebaseToken: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  error: string | null;

  // Actions
  initialize: () => void;
  loginWithGoogle: () => Promise<void>;
  loginWithGithub: () => Promise<void>;
  logout: () => Promise<void>;
  clearError: () => void;

  // Internal
  setFirebaseUser: (user: FirebaseUser | null) => void;
  setBackendUser: (user: BackendUser | null) => void;
  setFirebaseToken: (token: string | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  // Initial state
  firebaseUser: null,
  backendUser: null,
  firebaseToken: null,
  isLoading: true,
  isAuthenticated: false,
  error: null,

  // Initialize auth listener
  initialize: () => {
    console.log("[AuthStore] Initializing auth state listener");

    const unsubscribe = onAuthStateChanged(async (firebaseUser) => {
      console.log("[AuthStore] Auth state changed:", firebaseUser?.email || "null");

      set({ firebaseUser, isLoading: true });

      if (firebaseUser) {
        try {
          // Get Firebase ID token
          const token = await getIdToken();
          set({ firebaseToken: token });

          // Call /auth/me to get backend user (triggers auto-sync)
          const backendUser = await apiClient.getMe();
          console.log("[AuthStore] Backend user synced:", backendUser);

          set({
            backendUser,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          });
        } catch (error: any) {
          console.error("[AuthStore] Failed to sync backend user:", error);
          set({
            error: error.message || "Failed to sync user data",
            isLoading: false,
            isAuthenticated: false,
          });
        }
      } else {
        // User signed out
        set({
          firebaseUser: null,
          backendUser: null,
          firebaseToken: null,
          isAuthenticated: false,
          isLoading: false,
        });
      }
    });

    // Store unsubscribe function (we won't call it since we want persistent listener)
    // In a real app, you'd call this on unmount, but for our SPA it's fine to keep it
  },

  // Login with Google
  loginWithGoogle: async () => {
    set({ isLoading: true, error: null });
    try {
      await signInWithGoogle();
      // Auth state listener will handle the rest
    } catch (error: any) {
      console.error("[AuthStore] Google login failed:", error);
      set({
        error: error.message || "Failed to sign in with Google",
        isLoading: false,
      });
    }
  },

  // Login with GitHub
  loginWithGithub: async () => {
    set({ isLoading: true, error: null });
    try {
      await signInWithGithub();
      // Auth state listener will handle the rest
    } catch (error: any) {
      console.error("[AuthStore] GitHub login failed:", error);
      set({
        error: error.message || "Failed to sign in with GitHub",
        isLoading: false,
      });
    }
  },

  // Logout
  logout: async () => {
    set({ isLoading: true, error: null });
    try {
      // Call backend logout to revoke tokens
      await apiClient.logout();

      // Sign out from Firebase
      await firebaseSignOut();

      console.log("[AuthStore] Logged out successfully");
      // Auth state listener will clear the state
    } catch (error: any) {
      console.error("[AuthStore] Logout failed:", error);
      set({
        error: error.message || "Failed to log out",
        isLoading: false,
      });
    }
  },

  // Clear error
  clearError: () => set({ error: null }),

  // Internal setters
  setFirebaseUser: (firebaseUser) => set({ firebaseUser }),
  setBackendUser: (backendUser) => set({ backendUser }),
  setFirebaseToken: (firebaseToken) => set({ firebaseToken }),
  setLoading: (isLoading) => set({ isLoading }),
  setError: (error) => set({ error }),
}));
