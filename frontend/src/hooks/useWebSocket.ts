import { useEffect, useRef, useState } from "react";
import { useChatStore } from "@/stores/chatStore";
import { useAuthStore } from "@/stores/authStore";
import { getUserDisplayName, getUserColor } from "@/utils/user";
import { getSessionToken, clearSession } from "@/lib/session";

interface UserInfo {
  user_id: string;
  username?: string;
  color?: string;
}

interface WebSocketMessage {
  type: "chat" | "user_joined" | "user_left" | "user_sync" | "username_changed" | "color_changed";
  user_id: string;
  message_id?: string;
  payload?: string;
  position?: { x: number; y: number };
  users?: UserInfo[]; // For user_sync messages
  username?: string; // For user_joined and username_changed
  color?: string; // For user_joined and color_changed
  channel_id: string;
  timestamp: number;
}

interface UseWebSocketOptions {
  channelId?: string;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Event) => void;
}

export const useWebSocket = (options: UseWebSocketOptions = {}) => {
  const { channelId = "general", onConnect, onDisconnect, onError } = options;

  const socketRef = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [serverUserId, setServerUserId] = useState<string>(""); // Server-issued user ID
  const firebaseToken = useAuthStore((state) => state.firebaseToken);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const backendUser = useAuthStore((state) => state.backendUser);
  const addMessage = useChatStore((state) => state.addMessage);
  const addUser = useChatStore((state) => state.addUser);
  const updateUserUsername = useChatStore((state) => state.updateUserUsername);
  const updateUserColor = useChatStore((state) => state.updateUserColor);
  const removeUser = useChatStore((state) => state.removeUser);
  const clearUsers = useChatStore((state) => state.clearUsers);

  // Use refs for callbacks to avoid recreating the effect
  const addMessageRef = useRef(addMessage);
  const addUserRef = useRef(addUser);
  const updateUserUsernameRef = useRef(updateUserUsername);
  const updateUserColorRef = useRef(updateUserColor);
  const removeUserRef = useRef(removeUser);
  const clearUsersRef = useRef(clearUsers);
  const onConnectRef = useRef(onConnect);
  const onDisconnectRef = useRef(onDisconnect);
  const onErrorRef = useRef(onError);

  // Update refs when callbacks change
  useEffect(() => {
    addMessageRef.current = addMessage;
    addUserRef.current = addUser;
    updateUserUsernameRef.current = updateUserUsername;
    updateUserColorRef.current = updateUserColor;
    removeUserRef.current = removeUser;
    clearUsersRef.current = clearUsers;
    onConnectRef.current = onConnect;
    onDisconnectRef.current = onDisconnect;
    onErrorRef.current = onError;
  }, [addMessage, addUser, updateUserUsername, updateUserColor, removeUser, clearUsers, onConnect, onDisconnect, onError]);

  useEffect(() => {
    if (typeof window === "undefined") return;

    console.log("[WebSocket] Initializing connection...", {
      env: process.env.NODE_ENV,
      channelId,
    });

    // Async function to establish connection
    const connect = async () => {
      // Get token and user ID - either Firebase token or anonymous session token
      let token: string;
      let userId: string;

      if (firebaseToken && backendUser) {
        // Use Firebase token for authenticated users
        token = firebaseToken;
        userId = backendUser.id; // Use Firebase user ID

        // Clear anonymous session when switching to authenticated
        clearSession();
      } else {
        // Get anonymous session token
        const session = await getSessionToken();
        token = session.token;
        userId = session.userId; // Use server-issued anonymous user ID
      }

      // Store the server-issued user ID
      setServerUserId(userId);

      // Prioritize authenticated username over localStorage
      const username = isAuthenticated && backendUser
        ? backendUser.username
        : getUserDisplayName();
      const color = getUserColor();

      // Build query parameters - token is REQUIRED
      const params = new URLSearchParams({
        token,  // REQUIRED: Session or Firebase token
        room: channelId,
      });

      // Optional: username and color
      if (username) {
        params.set("username", username);
      }
      if (color) {
        params.set("color", color);
      }

      // Build WebSocket URL
      let wsUrl: string;
      if (process.env.NODE_ENV === "development") {
        const backendUrl = process.env.NEXT_PUBLIC_BACKEND_WS_URL || "ws://localhost:3001/api/chat";
        wsUrl = `${backendUrl}?${params.toString()}`;
      } else {
        // In production: use same host
        const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
        wsUrl = `${protocol}//${window.location.host}/api/chat?${params.toString()}`;
      }

      console.log("[WebSocket] Connecting to:", wsUrl.replace(/token=[^&]+/, 'token=***'));

      const socket = new WebSocket(wsUrl);
      socketRef.current = socket;

      socket.onopen = () => {
        console.log("[WebSocket] Connected successfully", { userId });
        setIsConnected(true);
        onConnectRef.current?.();
      };

      socket.onmessage = (event) => {
        console.log("[WebSocket] Message received");
        try {
          const data: WebSocketMessage = JSON.parse(event.data);

          // Handle different message types
          if (data.type === "user_sync") {
            // Initial sync of all users in channel
            console.log("[WebSocket] User sync:", data.users);
            if (data.users) {
              data.users.forEach((userInfo) => {
                if (typeof userInfo === "string") {
                  // Backwards compatibility: just user ID
                  addUserRef.current(userInfo);
                } else {
                  // New format: user object with username and color
                  addUserRef.current(userInfo.user_id, userInfo.username, userInfo.color);
                }
              });
            }
          } else if (data.type === "user_joined") {
            console.log("[WebSocket] User joined:", data.user_id, data.username, data.color);
            addUserRef.current(data.user_id, data.username, data.color);
          } else if (data.type === "username_changed") {
            console.log("[WebSocket] Username changed:", data.user_id, data.username);
            updateUserUsernameRef.current(data.user_id, data.username || "");
          } else if (data.type === "color_changed") {
            console.log("[WebSocket] Color changed:", data.user_id, data.color);
            updateUserColorRef.current(data.user_id, data.color || "");
          } else if (data.type === "user_left") {
            console.log("[WebSocket] User left:", data.user_id);
            removeUserRef.current(data.user_id);
          } else if (data.type === "chat") {
            // Handle chat message
            const { user_id, message_id, payload, position } = data;
            if (message_id && payload && position) {
              addMessageRef.current(message_id, user_id, payload, position.x, position.y);
            }
          }
        } catch (error) {
          console.error("[WebSocket] Failed to parse message:", error);
        }
      };

      socket.onerror = (error) => {
        console.error("[WebSocket] Error:", error);
        onErrorRef.current?.(error);
      };

      socket.onclose = (event) => {
        console.log("[WebSocket] Connection closed:", {
          code: event.code,
          reason: event.reason,
          wasClean: event.wasClean,
        });
        setIsConnected(false);
        onDisconnectRef.current?.();
      };
    };

    // Call connect async function
    connect().catch((err) => {
      console.error("[WebSocket] Failed to establish connection:", err);
      onErrorRef.current?.(err);
    });

    return () => {
      console.log("[WebSocket] 🧹 Cleanup - closing connection");
      if (socketRef.current) {
        const socket = socketRef.current;
        if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
          socket.close(1000, "Reconnecting");
        }
        // Clear the ref to prevent stale references
        socketRef.current = null;
      }
      // Clear users list to get fresh sync on reconnect
      clearUsersRef.current();
      // Reset connection state
      setIsConnected(false);
      setServerUserId("");
    };
  }, [channelId, firebaseToken, isAuthenticated, backendUser]); // Reconnect when channelId or auth state changes

  const sendMessage = (messageId: string, content: string, x: number, y: number) => {
    if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
      console.warn("WebSocket is not connected");
      return;
    }

    if (!serverUserId) {
      console.warn("Server user ID not yet available");
      return;
    }

    const message: WebSocketMessage = {
      type: "chat",
      user_id: serverUserId, // Use server-issued user ID
      message_id: messageId,
      payload: content,
      position: { x, y },
      channel_id: channelId,
      timestamp: Date.now(),
    };

    socketRef.current.send(JSON.stringify(message));
  };

  const sendUsernameChange = (username: string) => {
    if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
      console.warn("WebSocket is not connected");
      return;
    }

    if (!serverUserId) {
      console.warn("Server user ID not yet available");
      return;
    }

    const message: WebSocketMessage = {
      type: "username_changed",
      user_id: serverUserId, // Use server-issued user ID
      username,
      channel_id: channelId,
      timestamp: Date.now(),
    };

    socketRef.current.send(JSON.stringify(message));
  };

  const sendColorChange = (color: string) => {
    if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
      console.warn("WebSocket is not connected");
      return;
    }

    if (!serverUserId) {
      console.warn("Server user ID not yet available");
      return;
    }

    const message: WebSocketMessage = {
      type: "color_changed",
      user_id: serverUserId, // Use server-issued user ID
      color,
      channel_id: channelId,
      timestamp: Date.now(),
    };

    socketRef.current.send(JSON.stringify(message));
  };

  return {
    isConnected,
    sendMessage,
    sendUsernameChange,
    sendColorChange,
    serverUserId, // Expose server-issued user ID for components
  };
};
