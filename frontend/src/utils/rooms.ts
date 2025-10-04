/**
 * Room history management
 * Stores up to 10 most recently joined rooms in localStorage
 */

const USER_ROOMS_KEY = "asocial_user_rooms";
const MAX_ROOMS = 10;

/**
 * Get list of user's recently joined rooms
 * Returns empty array if no rooms found
 */
export function getUserRooms(): string[] {
  if (typeof window === "undefined") return [];

  const stored = localStorage.getItem(USER_ROOMS_KEY);
  if (!stored) return [];

  try {
    const rooms = JSON.parse(stored);
    return Array.isArray(rooms) ? rooms : [];
  } catch {
    return [];
  }
}

/**
 * Add a room to user's history
 * Moves room to front if already exists
 * Removes oldest room if at max capacity (10 rooms)
 */
export function addUserRoom(roomId: string): void {
  if (typeof window === "undefined") return;
  if (!roomId || roomId.trim() === "") return;

  const rooms = getUserRooms();

  // Remove room if it already exists (we'll add it to front)
  const filtered = rooms.filter((id) => id !== roomId);

  // Add to front
  const updated = [roomId, ...filtered];

  // Keep only last 10 rooms (FIFO)
  const trimmed = updated.slice(0, MAX_ROOMS);

  localStorage.setItem(USER_ROOMS_KEY, JSON.stringify(trimmed));
}

/**
 * Remove a room from user's history
 */
export function removeUserRoom(roomId: string): void {
  if (typeof window === "undefined") return;

  const rooms = getUserRooms();
  const filtered = rooms.filter((id) => id !== roomId);

  localStorage.setItem(USER_ROOMS_KEY, JSON.stringify(filtered));
}

/**
 * Clear all room history
 */
export function clearUserRooms(): void {
  if (typeof window === "undefined") return;
  localStorage.removeItem(USER_ROOMS_KEY);
}
