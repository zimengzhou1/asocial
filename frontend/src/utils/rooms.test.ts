import { describe, it, expect, beforeEach } from "vitest";
import {
  getUserRooms,
  addUserRoom,
  removeUserRoom,
  clearUserRooms,
} from "./rooms";

describe("Room utilities", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  describe("getUserRooms", () => {
    it("should return empty array when no rooms stored", () => {
      expect(getUserRooms()).toEqual([]);
    });

    it("should return array of stored rooms", () => {
      addUserRoom("room-1");
      addUserRoom("room-2");
      const rooms = getUserRooms();
      expect(rooms).toEqual(["room-2", "room-1"]);
    });

    it("should handle corrupted localStorage data", () => {
      localStorage.setItem("asocial_user_rooms", "invalid-json");
      expect(getUserRooms()).toEqual([]);
    });

    it("should handle non-array data in localStorage", () => {
      localStorage.setItem("asocial_user_rooms", JSON.stringify({ foo: "bar" }));
      expect(getUserRooms()).toEqual([]);
    });
  });

  describe("addUserRoom", () => {
    it("should add room to empty list", () => {
      addUserRoom("room-1");
      expect(getUserRooms()).toEqual(["room-1"]);
    });

    it("should add room to front of list", () => {
      addUserRoom("room-1");
      addUserRoom("room-2");
      expect(getUserRooms()).toEqual(["room-2", "room-1"]);
    });

    it("should move existing room to front", () => {
      addUserRoom("room-1");
      addUserRoom("room-2");
      addUserRoom("room-3");
      addUserRoom("room-1"); // Move to front
      expect(getUserRooms()).toEqual(["room-1", "room-3", "room-2"]);
    });

    it("should enforce 10 room limit (FIFO)", () => {
      // Add 12 rooms
      for (let i = 1; i <= 12; i++) {
        addUserRoom(`room-${i}`);
      }

      const rooms = getUserRooms();
      expect(rooms).toHaveLength(10);
      // Should have rooms 12 to 3 (newest to oldest)
      expect(rooms[0]).toBe("room-12");
      expect(rooms[9]).toBe("room-3");
      // rooms 1 and 2 should be removed
      expect(rooms).not.toContain("room-1");
      expect(rooms).not.toContain("room-2");
    });

    it("should ignore empty room IDs", () => {
      addUserRoom("");
      expect(getUserRooms()).toEqual([]);
    });

    it("should ignore whitespace-only room IDs", () => {
      addUserRoom("   ");
      expect(getUserRooms()).toEqual([]);
    });
  });

  describe("removeUserRoom", () => {
    it("should remove room from list", () => {
      addUserRoom("room-1");
      addUserRoom("room-2");
      removeUserRoom("room-1");
      expect(getUserRooms()).toEqual(["room-2"]);
    });

    it("should handle removing non-existent room", () => {
      addUserRoom("room-1");
      removeUserRoom("room-2");
      expect(getUserRooms()).toEqual(["room-1"]);
    });

    it("should handle removing from empty list", () => {
      expect(() => removeUserRoom("room-1")).not.toThrow();
      expect(getUserRooms()).toEqual([]);
    });

    it("should remove all instances of a room", () => {
      // Manually create duplicate (shouldn't happen with addUserRoom, but test defensively)
      localStorage.setItem(
        "asocial_user_rooms",
        JSON.stringify(["room-1", "room-2", "room-1"])
      );
      removeUserRoom("room-1");
      expect(getUserRooms()).toEqual(["room-2"]);
    });
  });

  describe("clearUserRooms", () => {
    it("should clear all rooms", () => {
      addUserRoom("room-1");
      addUserRoom("room-2");
      clearUserRooms();
      expect(getUserRooms()).toEqual([]);
    });

    it("should handle clearing when no rooms exist", () => {
      expect(() => clearUserRooms()).not.toThrow();
      expect(getUserRooms()).toEqual([]);
    });
  });

  describe("Room limit edge cases", () => {
    it("should maintain exactly 10 rooms at limit", () => {
      // Add exactly 10 rooms
      for (let i = 1; i <= 10; i++) {
        addUserRoom(`room-${i}`);
      }
      expect(getUserRooms()).toHaveLength(10);

      // Add one more
      addUserRoom("room-11");
      expect(getUserRooms()).toHaveLength(10);
      expect(getUserRooms()[0]).toBe("room-11");
      expect(getUserRooms()).not.toContain("room-1");
    });

    it("should handle moving room to front at limit", () => {
      // Add 10 rooms
      for (let i = 1; i <= 10; i++) {
        addUserRoom(`room-${i}`);
      }

      // Move room-5 to front (doesn't remove anything, just reorders)
      addUserRoom("room-5");

      const rooms = getUserRooms();
      expect(rooms).toHaveLength(10);
      expect(rooms[0]).toBe("room-5");
      // All rooms should still be present (just reordered)
      expect(rooms).toContain("room-1");
      expect(rooms).toContain("room-10");
    });
  });
});
