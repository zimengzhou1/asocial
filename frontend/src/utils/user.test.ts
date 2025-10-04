import { describe, it, expect, beforeEach } from "vitest";
import {
  getOrCreateUserId,
  getUserDisplayName,
  setUserDisplayName,
  clearUserData,
  getUserColor,
  setUserColor,
  COLOR_PALETTE,
} from "./user";

describe("User utilities", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  describe("getOrCreateUserId", () => {
    it("should create new user ID when none exists", () => {
      const userId = getOrCreateUserId();
      expect(userId).toBeTruthy();
      expect(typeof userId).toBe("string");
    });

    it("should return same user ID on subsequent calls", () => {
      const userId1 = getOrCreateUserId();
      const userId2 = getOrCreateUserId();
      expect(userId1).toBe(userId2);
    });

    it("should persist user ID in localStorage", () => {
      const userId = getOrCreateUserId();
      expect(localStorage.getItem("asocial_user_id")).toBe(userId);
    });
  });

  describe("getUserDisplayName / setUserDisplayName", () => {
    it("should return null when no username is stored", () => {
      expect(getUserDisplayName()).toBeNull();
    });

    it("should store and retrieve username", () => {
      const username = "Alice";
      setUserDisplayName(username);
      expect(getUserDisplayName()).toBe(username);
    });

    it("should update existing username", () => {
      setUserDisplayName("Alice");
      setUserDisplayName("Bob");
      expect(getUserDisplayName()).toBe("Bob");
    });
  });

  describe("getUserColor / setUserColor", () => {
    it("should return null when no color is stored", () => {
      expect(getUserColor()).toBeNull();
    });

    it("should store and retrieve valid color from palette", () => {
      const color = COLOR_PALETTE[0];
      setUserColor(color);
      expect(getUserColor()).toBe(color);
    });

    it("should reject colors not in palette", () => {
      const invalidColor = "#ffffff";
      setUserColor(invalidColor);
      expect(getUserColor()).toBeNull();
    });

    it("should store all valid palette colors", () => {
      COLOR_PALETTE.forEach((color) => {
        setUserColor(color);
        expect(getUserColor()).toBe(color);
      });
    });

    it("should update existing color", () => {
      setUserColor(COLOR_PALETTE[0]);
      setUserColor(COLOR_PALETTE[1]);
      expect(getUserColor()).toBe(COLOR_PALETTE[1]);
    });
  });

  describe("clearUserData", () => {
    it("should clear all user data", () => {
      getOrCreateUserId(); // Create a user ID
      setUserDisplayName("Alice");
      setUserColor(COLOR_PALETTE[0]);

      clearUserData();

      expect(localStorage.getItem("asocial_user_id")).toBeNull();
      expect(getUserDisplayName()).toBeNull();
      expect(getUserColor()).toBeNull();
    });

    it("should handle clearing when no data exists", () => {
      expect(() => clearUserData()).not.toThrow();
    });
  });

  describe("COLOR_PALETTE", () => {
    it("should contain 8 colors", () => {
      expect(COLOR_PALETTE).toHaveLength(8);
    });

    it("should contain valid hex colors", () => {
      COLOR_PALETTE.forEach((color) => {
        expect(color).toMatch(/^#[0-9a-f]{6}$/);
      });
    });

    it("should not contain duplicate colors", () => {
      const uniqueColors = new Set(COLOR_PALETTE);
      expect(uniqueColors.size).toBe(COLOR_PALETTE.length);
    });
  });
});
