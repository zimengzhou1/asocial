import { initializeApp, getApps, FirebaseApp } from "firebase/app";
import { getAuth, Auth } from "firebase/auth";

const firebaseConfig = {
  apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY,
  authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN,
  projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID,
};

// Validate configuration
function validateFirebaseConfig() {
  const missing = [];
  if (!firebaseConfig.apiKey) missing.push("NEXT_PUBLIC_FIREBASE_API_KEY");
  if (!firebaseConfig.authDomain) missing.push("NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN");
  if (!firebaseConfig.projectId) missing.push("NEXT_PUBLIC_FIREBASE_PROJECT_ID");

  if (missing.length > 0) {
    console.warn(
      `Firebase configuration incomplete. Missing: ${missing.join(", ")}. ` +
      `Authentication features will not work. Please add these to .env.local`
    );
    return false;
  }
  return true;
}

// Initialize Firebase (singleton pattern)
let app: FirebaseApp | null = null;
let auth: Auth | null = null;

export function initializeFirebase(): { app: FirebaseApp | null; auth: Auth | null } {
  // Return existing instance if already initialized
  if (app && auth) {
    return { app, auth };
  }

  // Validate config before initializing
  if (!validateFirebaseConfig()) {
    return { app: null, auth: null };
  }

  // Initialize only if not already initialized
  if (getApps().length === 0) {
    app = initializeApp(firebaseConfig);
    auth = getAuth(app);
    console.log("✅ Firebase initialized successfully");
  } else {
    app = getApps()[0];
    auth = getAuth(app);
  }

  return { app, auth };
}

// Get Firebase auth instance
export function getFirebaseAuth(): Auth | null {
  if (!auth) {
    const { auth: initializedAuth } = initializeFirebase();
    return initializedAuth;
  }
  return auth;
}

// Check if Firebase is properly configured
export function isFirebaseConfigured(): boolean {
  return validateFirebaseConfig();
}
