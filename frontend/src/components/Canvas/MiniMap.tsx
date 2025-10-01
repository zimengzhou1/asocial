"use client";

import React, { useState } from "react";

interface Message {
  id: string;
  userId: string;
  x: number;
  y: number;
  color: string;
}

interface MiniMapProps {
  messages: Message[];
  canvasWidth: number;
  canvasHeight: number;
  viewportX: number;
  viewportY: number;
  viewportScale: number;
  onJump: (x: number, y: number) => void;
}

const MiniMap: React.FC<MiniMapProps> = ({
  messages,
  canvasWidth,
  canvasHeight,
  viewportX,
  viewportY,
  viewportScale,
  onJump,
}) => {
  const [isOpen, setIsOpen] = useState(true);

  const miniMapWidth = 200;
  const miniMapHeight = 150;
  const scaleX = miniMapWidth / canvasWidth;
  const scaleY = miniMapHeight / canvasHeight;

  // Calculate viewport rectangle on minimap
  const viewportWidth = window.innerWidth / viewportScale;
  const viewportHeight = window.innerHeight / viewportScale;
  const viewportLeft = (-viewportX / viewportScale) * scaleX;
  const viewportTop = (-viewportY / viewportScale) * scaleY;
  const viewportRectWidth = viewportWidth * scaleX;
  const viewportRectHeight = viewportHeight * scaleY;

  const handleMiniMapClick = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const clickX = e.clientX - rect.left;
    const clickY = e.clientY - rect.top;

    // Convert minimap coordinates to canvas coordinates
    const canvasX = (clickX / scaleX);
    const canvasY = (clickY / scaleY);

    // Center the viewport on the clicked position
    const targetX = -canvasX * viewportScale + window.innerWidth / 2;
    const targetY = -canvasY * viewportScale + window.innerHeight / 2;

    onJump(targetX, targetY);
  };

  return (
    <>
      {!isOpen && (
        <div className="fixed top-4 right-4 z-50">
          <button
            onClick={() => setIsOpen(true)}
            className="p-2 bg-white rounded-lg shadow-lg hover:bg-gray-50 transition-colors border border-gray-300"
            aria-label="Open minimap"
          >
            <svg
              className="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7"
              />
            </svg>
          </button>
        </div>
      )}

      {isOpen && (
        <div className="fixed top-4 right-4 z-50 bg-white rounded-lg shadow-lg border border-gray-300 p-2">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-semibold text-gray-600">Map</span>
            <button
              onClick={() => setIsOpen(false)}
              className="p-1 hover:bg-gray-100 rounded transition-colors"
              aria-label="Close minimap"
            >
              <svg
                className="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          <div
            className="relative bg-gray-50 border border-gray-200 cursor-pointer rounded overflow-hidden"
            style={{
              width: `${miniMapWidth}px`,
              height: `${miniMapHeight}px`,
            }}
            onClick={handleMiniMapClick}
          >
            {/* Grid pattern */}
            <div
              className="absolute inset-0 opacity-30"
              style={{
                backgroundImage: `
                  linear-gradient(#d1d5db 1px, transparent 1px),
                  linear-gradient(90deg, #d1d5db 1px, transparent 1px)
                `,
                backgroundSize: `${20 * scaleX}px ${20 * scaleY}px`,
              }}
            />

            {/* Messages as dots */}
            {messages.map((msg) => (
              <div
                key={msg.id}
                className="absolute rounded-full"
                style={{
                  left: `${msg.x * scaleX}px`,
                  top: `${msg.y * scaleY}px`,
                  width: "4px",
                  height: "4px",
                  backgroundColor: msg.color,
                  transform: "translate(-50%, -50%)",
                }}
              />
            ))}

            {/* Viewport rectangle */}
            <div
              className="absolute border-2 border-blue-500 bg-blue-500 bg-opacity-10 pointer-events-none"
              style={{
                left: `${viewportLeft}px`,
                top: `${viewportTop}px`,
                width: `${viewportRectWidth}px`,
                height: `${viewportRectHeight}px`,
              }}
            />
          </div>
        </div>
      )}
    </>
  );
};

export default MiniMap;
