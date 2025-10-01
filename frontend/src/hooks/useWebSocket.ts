import { useEffect, useRef, useState } from "react";
import { useChatStore } from "@/stores/chatStore";

interface WebSocketMessage {
  type: "chat" | "user_joined" | "user_left" | "user_sync";
  user_id: string;
  message_id?: string;
  payload?: string;
  position?: { x: number; y: number };
  users?: string[]; // For user_sync messages
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
  const { channelId = "default", onConnect, onDisconnect, onError } = options;

  const socketRef = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const localUserId = useChatStore((state) => state.localUserId);
  const addMessage = useChatStore((state) => state.addMessage);
  const addUser = useChatStore((state) => state.addUser);
  const removeUser = useChatStore((state) => state.removeUser);

  // Use refs for callbacks to avoid recreating the effect
  const addMessageRef = useRef(addMessage);
  const addUserRef = useRef(addUser);
  const removeUserRef = useRef(removeUser);
  const onConnectRef = useRef(onConnect);
  const onDisconnectRef = useRef(onDisconnect);
  const onErrorRef = useRef(onError);

  // Update refs when callbacks change
  useEffect(() => {
    addMessageRef.current = addMessage;
    addUserRef.current = addUser;
    removeUserRef.current = removeUser;
    onConnectRef.current = onConnect;
    onDisconnectRef.current = onDisconnect;
    onErrorRef.current = onError;
  }, [addMessage, addUser, removeUser, onConnect, onDisconnect, onError]);

  useEffect(() => {
    if (typeof window === "undefined") return;

    console.log("[WebSocket] Initializing connection...", {
      env: process.env.NODE_ENV,
      localUserId,
      channelId,
    });

    // Determine WebSocket URL based on environment
    const getWebSocketUrl = () => {
      // In development: use env var or fallback to default
      if (process.env.NODE_ENV === "development") {
        const backendUrl = process.env.NEXT_PUBLIC_BACKEND_WS_URL || "ws://localhost:3001/api/chat";
        return `${backendUrl}?uid=${localUserId}`;
      }

      // In production (Docker): use same host, let Traefik proxy it
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      return `${protocol}//${window.location.host}/api/chat?uid=${localUserId}`;
    };

    const wsUrl = getWebSocketUrl();
    console.log("[WebSocket] Connecting to:", wsUrl);

    const socket = new WebSocket(wsUrl);
    socketRef.current = socket;

    socket.onopen = () => {
      console.log("[WebSocket] ✅ Connected successfully");
      setIsConnected(true);
      onConnectRef.current?.();
    };

    socket.onmessage = (event) => {
      console.log("[WebSocket] 📨 Message received");
      try {
        const data: WebSocketMessage = JSON.parse(event.data);

        // Handle different message types
        if (data.type === "user_sync") {
          // Initial sync of all users in channel
          console.log("[WebSocket] 🔄 User sync:", data.users);
          if (data.users) {
            data.users.forEach((userId) => addUserRef.current(userId));
          }
        } else if (data.type === "user_joined") {
          console.log("[WebSocket] 👋 User joined:", data.user_id);
          addUserRef.current(data.user_id);
        } else if (data.type === "user_left") {
          console.log("[WebSocket] 👋 User left:", data.user_id);
          removeUserRef.current(data.user_id);
        } else if (data.type === "chat") {
          // Handle chat message
          const { user_id, message_id, payload, position } = data;
          if (message_id && payload && position) {
            addMessageRef.current(message_id, user_id, payload, position.x, position.y);
          }
        }
      } catch (error) {
        console.error("[WebSocket] ❌ Failed to parse message:", error);
      }
    };

    socket.onerror = (error) => {
      console.error("[WebSocket] ❌ Error:", error);
      onErrorRef.current?.(error);
    };

    socket.onclose = (event) => {
      console.log("[WebSocket] ⚠️ Connection closed:", {
        code: event.code,
        reason: event.reason,
        wasClean: event.wasClean,
      });
      setIsConnected(false);
      onDisconnectRef.current?.();
    };

    return () => {
      console.log("[WebSocket] 🧹 Cleanup - closing connection");
      if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
        socket.close(1000, "Component unmounting");
      }
    };
  }, [localUserId, channelId]); // Only reconnect when userId or channelId changes

  const sendMessage = (messageId: string, content: string, x: number, y: number) => {
    if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
      console.warn("WebSocket is not connected");
      return;
    }

    const message: WebSocketMessage = {
      type: "chat",
      user_id: localUserId,
      message_id: messageId,
      payload: content,
      position: { x, y },
      channel_id: channelId,
      timestamp: Date.now(),
    };

    socketRef.current.send(JSON.stringify(message));
  };

  return {
    isConnected,
    sendMessage,
  };
};
