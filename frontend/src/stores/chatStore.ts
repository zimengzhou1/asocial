import { create } from "zustand";
import { generateUUID } from "@/utils/uuid";

const REMOVE_DELAY = 5000;

// Generate random color for user
const generateColor = (userId: string): string => {
  const colors = [
    "#ef4444", "#f59e0b", "#10b981", "#3b82f6",
    "#8b5cf6", "#ec4899", "#06b6d4", "#84cc16"
  ];
  let hash = 0;
  for (let i = 0; i < userId.length; i++) {
    hash = userId.charCodeAt(i) + ((hash << 5) - hash);
  }
  return colors[Math.abs(hash) % colors.length];
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
  addUser: (userId: string) => void;
  removeUser: (userId: string) => void;

  // Actions - Viewport
  setViewport: (viewport: Viewport) => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  // Initial state
  messages: {},
  users: {},
  viewport: { x: 0, y: 0, scale: 1 },
  localUserId: typeof window !== "undefined" ? generateUUID() : "",

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
  addUser: (userId) => {
    set((state) => {
      if (state.users[userId]) return state;

      return {
        users: {
          ...state.users,
          [userId]: {
            id: userId,
            color: generateColor(userId),
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
