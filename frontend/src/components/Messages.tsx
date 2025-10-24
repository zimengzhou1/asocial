"use client";

import React from "react";
import CanvasViewport from "@/components/Canvas/CanvasViewport";
import TextMessage from "@/components/Canvas/TextMessage";
import Sidebar from "@/components/Layout/Sidebar";
import MiniMap from "@/components/Canvas/MiniMap";
import { useChatStore } from "@/stores/chatStore";
import { useWebSocket } from "@/hooks/useWebSocket";
import { generateUUID } from "@/utils/uuid";

interface MessagesProps {
  roomSlug: string;
}

const Messages: React.FC<MessagesProps> = ({ roomSlug }) => {
  // Get state from Zustand store
  const messages = useChatStore((state) => state.messages);
  const users = useChatStore((state) => state.users);
  const viewport = useChatStore((state) => state.viewport);
  const setViewport = useChatStore((state) => state.setViewport);
  const addMessage = useChatStore((state) => state.addMessage);
  const updateMessage = useChatStore((state) => state.updateMessage);

  // Initialize WebSocket connection and get server-issued user ID
  const { sendMessage, sendUsernameChange, sendColorChange, serverUserId } =
    useWebSocket({
      channelId: roomSlug,
      onConnect: () => console.log("Connected to chat server"),
      onDisconnect: () => console.log("Disconnected from chat server"),
    });

  const handleCanvasClick = (x: number, y: number) => {
    if (!serverUserId) {
      console.warn("Cannot create message: server user ID not yet available");
      return;
    }
    const messageId = generateUUID();
    addMessage(messageId, serverUserId, "", x, y); // Use server-issued user ID
  };

  const handleContentChange = (messageId: string, content: string) => {
    const message = messages[messageId];
    if (!message) return;

    // Update store
    updateMessage(messageId, content);

    // Send to server
    sendMessage(messageId, content, message.x, message.y);
  };

  const handleRecenter = () => {
    // Calculate center position based on window size
    // Account for sidebar: left on desktop (56px), bottom on mobile (64px)
    const isMobile = window.innerWidth < 768; // md breakpoint
    const sidebarOffset = isMobile ? 0 : 56; // Left offset for desktop
    const bottomOffset = isMobile ? 64 : 0; // Bottom offset for mobile

    const centerX = window.innerWidth / 2 - 5000 / 2 - sidebarOffset;
    const centerY = window.innerHeight / 2 - 5000 / 2 - bottomOffset / 2;
    setViewport({ x: centerX, y: centerY, scale: 1 });
  };

  const handleViewportChange = (x: number, y: number, scale: number) => {
    setViewport({ x, y, scale });
  };

  const handleMiniMapJump = (x: number, y: number) => {
    setViewport({ x, y, scale: viewport.scale });
  };

  return (
    <div className="relative w-screen h-screen overflow-hidden">
      {/* Sidebar */}
      <Sidebar
        users={Object.values(users)}
        currentUserId={serverUserId || ""}
        onRecenter={handleRecenter}
        onUsernameChange={sendUsernameChange}
        onColorChange={sendColorChange}
      />

      {/* Mini Map */}
      <MiniMap
        messages={Object.values(messages)}
        canvasWidth={5000}
        canvasHeight={5000}
        viewportX={viewport.x}
        viewportY={viewport.y}
        viewportScale={viewport.scale}
        onJump={handleMiniMapJump}
      />

      {/* Canvas with messages - padding left on desktop, bottom on mobile */}
      <div className="absolute inset-0 pb-16 md:pb-0 md:pl-14 z-0">
        <CanvasViewport
          canvasWidth={5000}
          canvasHeight={5000}
          viewport={viewport}
          onCanvasClick={handleCanvasClick}
          onViewportChange={handleViewportChange}
        >
          {Object.values(messages).map((message) => (
            <TextMessage
              key={message.id}
              id={message.id}
              userId={message.userId}
              content={message.content}
              x={message.x}
              y={message.y}
              color={message.color}
              isLocal={message.userId === serverUserId}
              fadeOut={message.fadeOut}
              onContentChange={handleContentChange}
            />
          ))}
        </CanvasViewport>
      </div>
    </div>
  );
};

export default Messages;
