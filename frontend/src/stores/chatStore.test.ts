import { describe, it, expect, beforeEach, vi, afterEach } from "vitest";
import { useChatStore } from "./chatStore";

describe("ChatStore", () => {
  beforeEach(() => {
    // Reset store state before each test
    useChatStore.setState({
      messages: {},
      users: {},
      viewport: { x: 0, y: 0, scale: 1 },
    });

    // Clear timers
    vi.clearAllTimers();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("User actions", () => {
    it("should add user with generated color when no color provided", () => {
      const { addUser, users } = useChatStore.getState();

      addUser("user-123");

      const updatedUsers = useChatStore.getState().users;
      expect(updatedUsers["user-123"]).toBeDefined();
      expect(updatedUsers["user-123"].id).toBe("user-123");
      expect(updatedUsers["user-123"].color).toBeTruthy();
    });

    it("should add user with provided color", () => {
      const { addUser } = useChatStore.getState();

      addUser("user-123", "Alice", "#ef4444");

      const users = useChatStore.getState().users;
      expect(users["user-123"]).toEqual({
        id: "user-123",
        username: "Alice",
        color: "#ef4444",
      });
    });

    it("should add user with username but no color", () => {
      const { addUser } = useChatStore.getState();

      addUser("user-123", "Alice");

      const users = useChatStore.getState().users;
      expect(users["user-123"].username).toBe("Alice");
      expect(users["user-123"].color).toBeTruthy(); // Should have generated color
    });

    it("should not overwrite existing user", () => {
      const { addUser } = useChatStore.getState();

      addUser("user-123", "Alice", "#ef4444");
      addUser("user-123", "Bob", "#10b981"); // Try to overwrite

      const users = useChatStore.getState().users;
      expect(users["user-123"].username).toBe("Alice"); // Should remain Alice
      expect(users["user-123"].color).toBe("#ef4444");
    });

    it("should update user username", () => {
      const { addUser, updateUserUsername } = useChatStore.getState();

      addUser("user-123", "Alice", "#ef4444");
      updateUserUsername("user-123", "Alice Smith");

      const users = useChatStore.getState().users;
      expect(users["user-123"].username).toBe("Alice Smith");
      expect(users["user-123"].color).toBe("#ef4444"); // Color should remain
    });

    it("should not update username for non-existent user", () => {
      const { updateUserUsername } = useChatStore.getState();

      updateUserUsername("user-999", "Ghost");

      const users = useChatStore.getState().users;
      expect(users["user-999"]).toBeUndefined();
    });

    it("should update user color", () => {
      const { addUser, updateUserColor } = useChatStore.getState();

      addUser("user-123", "Alice", "#ef4444");
      updateUserColor("user-123", "#10b981");

      const users = useChatStore.getState().users;
      expect(users["user-123"].color).toBe("#10b981");
      expect(users["user-123"].username).toBe("Alice"); // Username should remain
    });

    it("should not update color for non-existent user", () => {
      const { updateUserColor } = useChatStore.getState();

      updateUserColor("user-999", "#ef4444");

      const users = useChatStore.getState().users;
      expect(users["user-999"]).toBeUndefined();
    });

    it("should remove user", () => {
      const { addUser, removeUser } = useChatStore.getState();

      addUser("user-123", "Alice", "#ef4444");
      addUser("user-456", "Bob", "#10b981");

      removeUser("user-123");

      const users = useChatStore.getState().users;
      expect(users["user-123"]).toBeUndefined();
      expect(users["user-456"]).toBeDefined();
    });

    it("should handle removing non-existent user", () => {
      const { addUser, removeUser } = useChatStore.getState();

      addUser("user-123", "Alice", "#ef4444");
      removeUser("user-999");

      const users = useChatStore.getState().users;
      expect(users["user-123"]).toBeDefined();
    });
  });

  describe("Message actions", () => {
    it("should add message", () => {
      const { addUser, addMessage } = useChatStore.getState();

      addUser("user-123", "Alice", "#ef4444");
      addMessage("msg-1", "user-123", "Hello World", 100, 200);

      const messages = useChatStore.getState().messages;
      expect(messages["msg-1"]).toBeDefined();
      expect(messages["msg-1"].content).toBe("Hello World");
      expect(messages["msg-1"].x).toBe(100);
      expect(messages["msg-1"].y).toBe(200);
      expect(messages["msg-1"].color).toBe("#ef4444");
      expect(messages["msg-1"].fadeOut).toBe(false);
    });

    it("should auto-add user if not exists when adding message", () => {
      const { addMessage } = useChatStore.getState();

      addMessage("msg-1", "user-123", "Hello", 0, 0);

      const users = useChatStore.getState().users;
      expect(users["user-123"]).toBeDefined();
    });

    it("should update message content", () => {
      const { addUser, addMessage, updateMessage } = useChatStore.getState();

      addUser("user-123", "Alice", "#ef4444");
      addMessage("msg-1", "user-123", "Hello", 100, 200);
      updateMessage("msg-1", "Hello World");

      const messages = useChatStore.getState().messages;
      expect(messages["msg-1"].content).toBe("Hello World");
      expect(messages["msg-1"].x).toBe(100); // Position should remain
    });

    it("should not update non-existent message", () => {
      const { updateMessage } = useChatStore.getState();

      expect(() => updateMessage("msg-999", "Ghost message")).not.toThrow();

      const messages = useChatStore.getState().messages;
      expect(messages["msg-999"]).toBeUndefined();
    });

    it("should fade out message", () => {
      const { addUser, addMessage, fadeOutMessage } = useChatStore.getState();

      addUser("user-123", "Alice", "#ef4444");
      addMessage("msg-1", "user-123", "Hello", 100, 200);

      fadeOutMessage("msg-1");

      const messages = useChatStore.getState().messages;
      expect(messages["msg-1"].fadeOut).toBe(true);
    });

    it("should not fade out non-existent message", () => {
      const { fadeOutMessage } = useChatStore.getState();

      expect(() => fadeOutMessage("msg-999")).not.toThrow();
    });

    it("should remove message", () => {
      const { addUser, addMessage, removeMessage } = useChatStore.getState();

      addUser("user-123", "Alice", "#ef4444");
      addMessage("msg-1", "user-123", "Hello", 100, 200);
      addMessage("msg-2", "user-123", "World", 150, 250);

      removeMessage("msg-1");

      const messages = useChatStore.getState().messages;
      expect(messages["msg-1"]).toBeUndefined();
      expect(messages["msg-2"]).toBeDefined();
    });

    it("should handle removing non-existent message", () => {
      const { addUser, addMessage, removeMessage } = useChatStore.getState();

      addUser("user-123", "Alice", "#ef4444");
      addMessage("msg-1", "user-123", "Hello", 100, 200);

      removeMessage("msg-999");

      const messages = useChatStore.getState().messages;
      expect(messages["msg-1"]).toBeDefined();
    });
  });

  describe("Viewport actions", () => {
    it("should set viewport", () => {
      const { setViewport } = useChatStore.getState();

      setViewport({ x: 100, y: 200, scale: 1.5 });

      const viewport = useChatStore.getState().viewport;
      expect(viewport).toEqual({ x: 100, y: 200, scale: 1.5 });
    });

    it("should update viewport independently", () => {
      const { setViewport } = useChatStore.getState();

      setViewport({ x: 50, y: 50, scale: 1 });
      setViewport({ x: 100, y: 100, scale: 2 });

      const viewport = useChatStore.getState().viewport;
      expect(viewport).toEqual({ x: 100, y: 100, scale: 2 });
    });
  });

  describe("Initial state", () => {
    it("should have correct initial state", () => {
      const state = useChatStore.getState();

      expect(state.messages).toEqual({});
      expect(state.users).toEqual({});
      expect(state.viewport).toEqual({ x: 0, y: 0, scale: 1 });
      expect(state.localUserId).toBeTruthy(); // Should be generated
    });
  });

  describe("Color generation", () => {
    it("should generate consistent color for same user ID", () => {
      const { addUser } = useChatStore.getState();

      // Clear and re-add user
      addUser("user-123");
      const color1 = useChatStore.getState().users["user-123"].color;

      useChatStore.setState({ users: {} });
      addUser("user-123");
      const color2 = useChatStore.getState().users["user-123"].color;

      expect(color1).toBe(color2);
    });

    it("should generate different colors for different users", () => {
      const { addUser } = useChatStore.getState();

      addUser("user-123");
      addUser("user-456");

      const color1 = useChatStore.getState().users["user-123"].color;
      const color2 = useChatStore.getState().users["user-456"].color;

      // Note: This might occasionally fail if hash collision occurs,
      // but it's very unlikely with only 2 users
      expect(color1).not.toBe(color2);
    });
  });
});
