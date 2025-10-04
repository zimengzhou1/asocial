import { generateUUID } from "./uuid";

// localStorage keys
const USER_ID_KEY = "asocial_user_id";
const USERNAME_KEY = "asocial_username";
const USER_COLOR_KEY = "asocial_user_color";

// Available color palette (same as chatStore)
export const COLOR_PALETTE = [
  "#ef4444", "#f59e0b", "#10b981", "#3b82f6",
  "#8b5cf6", "#ec4899", "#06b6d4", "#84cc16"
] as const;

/**
 * Get or create a persistent user ID
 * User ID is stored in localStorage and persists across sessions
 * This becomes the permanent identifier that can later be linked to an account
 */
export function getOrCreateUserId(): string {
  if (typeof window === "undefined") {
    return ""; // SSR guard
  }

  // Check if user ID already exists in localStorage
  let userId = localStorage.getItem(USER_ID_KEY);

  if (!userId) {
    // Generate new user ID and save it
    userId = generateUUID();
    localStorage.setItem(USER_ID_KEY, userId);
  }

  return userId;
}

/**
 * Get user's display name from localStorage
 * Returns null if no username is set
 */
export function getUserDisplayName(): string | null {
  if (typeof window === "undefined") {
    return null;
  }

  return localStorage.getItem(USERNAME_KEY);
}

/**
 * Set user's display name in localStorage
 */
export function setUserDisplayName(name: string): void {
  if (typeof window === "undefined") {
    return;
  }

  if (name.trim()) {
    localStorage.setItem(USERNAME_KEY, name.trim());
  } else {
    // Clear username if empty
    localStorage.removeItem(USERNAME_KEY);
  }
}

/**
 * Get user's color from localStorage
 * Returns null if no color is set
 */
export function getUserColor(): string | null {
  if (typeof window === "undefined") {
    return null;
  }

  return localStorage.getItem(USER_COLOR_KEY);
}

/**
 * Set user's color in localStorage
 * Only accepts colors from the predefined palette
 */
export function setUserColor(color: string): void {
  if (typeof window === "undefined") {
    return;
  }

  // Validate color is in palette
  if (COLOR_PALETTE.includes(color as any)) {
    localStorage.setItem(USER_COLOR_KEY, color);
  }
}

/**
 * Clear all user data from localStorage
 * Useful for "logout" or "reset identity" functionality
 */
export function clearUserData(): void {
  if (typeof window === "undefined") {
    return;
  }

  localStorage.removeItem(USER_ID_KEY);
  localStorage.removeItem(USERNAME_KEY);
  localStorage.removeItem(USER_COLOR_KEY);
}
