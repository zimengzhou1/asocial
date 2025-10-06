import { getIdToken } from "./auth";

const API_BASE_URL = process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:3001";

/**
 * User data from backend
 */
export interface BackendUser {
  id: string;
  email: string;
  username: string;
  created_at: string;
}

/**
 * Room data from backend
 */
export interface Room {
  id: string;
  name: string;
  slug: string;
  description?: string;
  is_public: boolean;
  created_at: string;
}

/**
 * Room user settings
 */
export interface RoomUserSettings {
  display_name: string;
  color: string;
  joined_at: string;
}

/**
 * Join room response
 */
export interface JoinRoomResponse {
  room: Room;
  settings: RoomUserSettings;
}

/**
 * API client for backend requests
 * Automatically includes Firebase ID token in Authorization header
 */
class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  /**
   * Make authenticated request to backend
   */
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const token = await getIdToken();

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    };

    // Add Bearer token if available
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(
        `API Error: ${response.status} ${response.statusText} - ${errorText}`
      );
    }

    return response.json();
  }

  /**
   * Get current user info from backend
   * Triggers auto-sync on first call after Firebase login
   */
  async getMe(): Promise<BackendUser> {
    return this.request<BackendUser>("/api/auth/me");
  }

  /**
   * Check if username is available
   */
  async checkUsername(username: string): Promise<{
    available: boolean;
    suggestions?: string[];
  }> {
    return this.request("/api/auth/check-username", {
      method: "POST",
      body: JSON.stringify({ username }),
    });
  }

  /**
   * Logout (revoke Firebase tokens on backend)
   */
  async logout(): Promise<void> {
    try {
      await this.request("/api/auth/logout", {
        method: "POST",
      });
    } catch (error) {
      console.error("Logout API error:", error);
      // Don't throw - logout should succeed even if API fails
    }
  }

  /**
   * Update username
   */
  async updateUsername(username: string): Promise<BackendUser> {
    return this.request<BackendUser>("/api/auth/username", {
      method: "PATCH",
      body: JSON.stringify({ username }),
    });
  }

  /**
   * Delete user account
   */
  async deleteAccount(): Promise<void> {
    await this.request("/api/auth/account", {
      method: "DELETE",
    });
  }

  /**
   * Join a room
   */
  async joinRoom(
    slug: string,
    displayName?: string,
    color?: string
  ): Promise<JoinRoomResponse> {
    const body: { display_name?: string; color?: string } = {};
    if (displayName) body.display_name = displayName;
    if (color) body.color = color;

    return this.request<JoinRoomResponse>(`/api/rooms/${slug}/join`, {
      method: "POST",
      body: JSON.stringify(body),
    });
  }

  /**
   * Get room by slug
   */
  async getRoom(slug: string): Promise<Room> {
    return this.request<Room>(`/api/rooms/${slug}`);
  }

  /**
   * List all public rooms
   */
  async listPublicRooms(): Promise<{ rooms: Room[]; count: number }> {
    return this.request<{ rooms: Room[]; count: number }>("/api/rooms/public");
  }
}

// Export singleton instance
export const apiClient = new ApiClient(API_BASE_URL);
