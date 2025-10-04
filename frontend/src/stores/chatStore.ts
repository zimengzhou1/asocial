import { create } from "zustand";
import { generateUUID } from "@/utils/uuid";
import { getOrCreateUserId, COLOR_PALETTE } from "@/utils/user";

const REMOVE_DELAY = 5000;

// Generate random color for user (fallback when no color provided)
const generateColor = (userId: string): string => {
  let hash = 0;
  for (let i = 0; i < userId.length; i++) {
    hash = userId.charCodeAt(i) + ((hash << 5) - hash);
  }
  return COLOR_PALETTE[Math.abs(hash) % COLOR_PALETTE.length];
};

export interface Message {
  id: string;
  userId: string;
  content: string;
  x: number;
  y: number;
  color: string;
  fadeOut: boolean;
  timeoutID: number;
}

export interface User {
  id: string;
  color: string;
  username?: string;
}

interface Viewport {
  x: number;
  y: number;
  scale: number;
}

interface ChatState {
  // State
  messages: { [key: string]: Message };
  users: { [key: string]: User };
  viewport: Viewport;
  localUserId: string;

  // Actions - Messages
  addMessage: (messageId: string, userId: string, content: string, x: number, y: number) => void;
  updateMessage: (messageId: string, content: string) => void;
  removeMessage: (messageId: string) => void;
  fadeOutMessage: (messageId: string) => void;

  // Actions - Users
  addUser: (userId: string, username?: string, color?: string) => void;
  updateUserUsername: (userId: string, username: string) => void;
  updateUserColor: (userId: string, color: string) => void;
  removeUser: (userId: string) => void;

  // Actions - Viewport
  setViewport: (viewport: Viewport) => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  // Initial state
  messages: {},
  users: {},
  viewport: { x: 0, y: 0, scale: 1 },
  localUserId: typeof window !== "undefined" ? getOrCreateUserId() : "",

  // Message actions
  addMessage: (messageId, userId, content, x, y) => {
    const state = get();

    // Add user if doesn't exist
    if (!state.users[userId]) {
      state.addUser(userId);
    }

    // Clear old timeout if message exists
    const existingMessage = state.messages[messageId];
    if (existingMessage) {
      window.clearTimeout(existingMessage.timeoutID);
    }

    const userColor = state.users[userId]?.color || generateColor(userId);
    const timeoutID = window.setTimeout(() => {
      state.fadeOutMessage(messageId);
    }, REMOVE_DELAY);

    set((state) => ({
      messages: {
        ...state.messages,
        [messageId]: {
          id: messageId,
          userId,
          content,
          x,
          y,
          color: userColor,
          fadeOut: false,
          timeoutID,
        },
      },
    }));
  },

  updateMessage: (messageId, content) => {
    const state = get();
    const message = state.messages[messageId];

    if (!message) return;

    // Clear old timeout and create new one
    window.clearTimeout(message.timeoutID);
    const newTimeoutID = window.setTimeout(() => {
      state.fadeOutMessage(messageId);
    }, REMOVE_DELAY);

    set((state) => ({
      messages: {
        ...state.messages,
        [messageId]: {
          ...state.messages[messageId],
          content,
          timeoutID: newTimeoutID,
        },
      },
    }));
  },

  fadeOutMessage: (messageId) => {
    const state = get();
    const message = state.messages[messageId];

    if (!message) return;

    window.clearTimeout(message.timeoutID);

    set((state) => ({
      messages: {
        ...state.messages,
        [messageId]: {
          ...state.messages[messageId],
          fadeOut: true,
        },
      },
    }));

    // Remove after fade animation
    setTimeout(() => {
      state.removeMessage(messageId);
    }, 500);
  },

  removeMessage: (messageId) => {
    set((state) => {
      const { [messageId]: removed, ...rest } = state.messages;
      return { messages: rest };
    });
  },

  // User actions
  addUser: (userId, username, color) => {
    set((state) => {
      if (state.users[userId]) return state;

      return {
        users: {
          ...state.users,
          [userId]: {
            id: userId,
            color: color || generateColor(userId), // Use provided color or generate
            username,
          },
        },
      };
    });
  },

  updateUserUsername: (userId, username) => {
    set((state) => {
      if (!state.users[userId]) return state;

      return {
        users: {
          ...state.users,
          [userId]: {
            ...state.users[userId],
            username,
          },
        },
      };
    });
  },

  updateUserColor: (userId, color) => {
    set((state) => {
      if (!state.users[userId]) return state;

      return {
        users: {
          ...state.users,
          [userId]: {
            ...state.users[userId],
            color,
          },
        },
      };
    });
  },

  removeUser: (userId) => {
    set((state) => {
      const { [userId]: removed, ...rest } = state.users;
      return { users: rest };
    });
  },

  // Viewport actions
  setViewport: (viewport) => {
    set({ viewport });
  },
}));
