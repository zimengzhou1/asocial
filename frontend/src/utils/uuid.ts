/**
 * Generate a UUID v4 string
 * Uses crypto.randomUUID() if available, falls back to custom implementation
 * for older mobile browsers (iOS Safari < 15.4, etc.)
 */
export function generateUUID(): string {
  // Use native crypto.randomUUID if available
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }

  // Fallback implementation for older browsers
  // Based on the RFC4122 version 4 UUID format
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}
