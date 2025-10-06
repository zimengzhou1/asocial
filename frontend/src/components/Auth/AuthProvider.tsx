"use client";

import { useEffect } from "react";
import { useAuthStore } from "@/stores/authStore";
import { initializeFirebase } from "@/lib/firebase";

/**
 * AuthProvider component
 * Initializes Firebase and sets up auth state listener
 * Place this high in the component tree (e.g., in layout)
 */
export default function AuthProvider({ children }: { children: React.ReactNode }) {
  const initialize = useAuthStore((state) => state.initialize);

  useEffect(() => {
    // Initialize Firebase
    initializeFirebase();

    // Initialize auth state listener
    initialize();
  }, [initialize]);

  return <>{children}</>;
}
