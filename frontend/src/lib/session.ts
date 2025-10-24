import { v4 as uuidv4 } from "uuid";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:3001";

interface SessionToken {
  token: string;
  userId: string;
  expiresAt: number;
}

const SESSION_TOKEN_KEY = "session_token";
const CLIENT_ID_KEY = "client_id";

/**
 * Get or create a session token for anonymous users
 */
export async function getSessionToken(): Promise<SessionToken> {
  // Check if we have a valid cached token
  const cached = getCachedToken();
  if (cached && cached.expiresAt > Date.now()) {
    return cached;
  }

  // Get or create client ID (persists across sessions)
  let clientId = localStorage.getItem(CLIENT_ID_KEY);
  if (!clientId) {
    clientId = uuidv4();
    localStorage.setItem(CLIENT_ID_KEY, clientId);
  }

  // Request new session token from backend
  const response = await fetch(`${API_BASE_URL}/api/auth/anonymous-session`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ client_id: clientId }),
  });

  if (!response.ok) {
    throw new Error("Failed to get session token");
  }

  const data = await response.json();
  const sessionToken: SessionToken = {
    token: data.session_token,
    userId: data.user_id,
    expiresAt: data.expires_at * 1000, // Convert to ms
  };

  // Cache token
  localStorage.setItem(SESSION_TOKEN_KEY, JSON.stringify(sessionToken));

  return sessionToken;
}

function getCachedToken(): SessionToken | null {
  const cached = localStorage.getItem(SESSION_TOKEN_KEY);
  if (!cached) return null;

  try {
    return JSON.parse(cached);
  } catch {
    return null;
  }
}

export function clearSession() {
  localStorage.removeItem(SESSION_TOKEN_KEY);
  // Keep client_id for session continuity
}
