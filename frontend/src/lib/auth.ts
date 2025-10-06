import {
  signInWithPopup,
  GoogleAuthProvider,
  GithubAuthProvider,
  signOut as firebaseSignOut,
  onAuthStateChanged as firebaseOnAuthStateChanged,
  User,
  UserCredential,
} from "firebase/auth";
import { getFirebaseAuth } from "./firebase";

const googleProvider = new GoogleAuthProvider();
const githubProvider = new GithubAuthProvider();

/**
 * Sign in with Google OAuth
 */
export async function signInWithGoogle(): Promise<UserCredential> {
  const auth = getFirebaseAuth();
  if (!auth) {
    throw new Error("Firebase not initialized. Please check your configuration.");
  }

  try {
    const result = await signInWithPopup(auth, googleProvider);
    console.log("✅ Signed in with Google:", result.user.email);
    return result;
  } catch (error: any) {
    console.error("❌ Google sign-in error:", error);
    throw new Error(`Failed to sign in with Google: ${error.message}`);
  }
}

/**
 * Sign in with GitHub OAuth
 */
export async function signInWithGithub(): Promise<UserCredential> {
  const auth = getFirebaseAuth();
  if (!auth) {
    throw new Error("Firebase not initialized. Please check your configuration.");
  }

  try {
    const result = await signInWithPopup(auth, githubProvider);
    console.log("✅ Signed in with GitHub:", result.user.email);
    return result;
  } catch (error: any) {
    console.error("❌ GitHub sign-in error:", error);
    throw new Error(`Failed to sign in with GitHub: ${error.message}`);
  }
}

/**
 * Sign out the current user
 */
export async function signOut(): Promise<void> {
  const auth = getFirebaseAuth();
  if (!auth) {
    throw new Error("Firebase not initialized.");
  }

  try {
    await firebaseSignOut(auth);
    console.log("✅ Signed out successfully");
  } catch (error: any) {
    console.error("❌ Sign-out error:", error);
    throw new Error(`Failed to sign out: ${error.message}`);
  }
}

/**
 * Get the current Firebase ID token
 * Returns null if user is not authenticated
 */
export async function getIdToken(): Promise<string | null> {
  const auth = getFirebaseAuth();
  if (!auth || !auth.currentUser) {
    return null;
  }

  try {
    const token = await auth.currentUser.getIdToken();
    return token;
  } catch (error: any) {
    console.error("❌ Failed to get ID token:", error);
    return null;
  }
}

/**
 * Subscribe to auth state changes
 * Returns unsubscribe function
 */
export function onAuthStateChanged(callback: (user: User | null) => void): () => void {
  const auth = getFirebaseAuth();
  if (!auth) {
    console.warn("Firebase not initialized, auth state listener not attached");
    return () => {}; // Return no-op unsubscribe
  }

  return firebaseOnAuthStateChanged(auth, callback);
}

/**
 * Get current Firebase user
 */
export function getCurrentUser(): User | null {
  const auth = getFirebaseAuth();
  return auth?.currentUser || null;
}
