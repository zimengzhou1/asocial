"use client";

import React, { useRef, useState, useEffect } from "react";
import { useGesture } from "@use-gesture/react";

interface CanvasViewportProps {
  children: React.ReactNode;
  canvasWidth?: number;
  canvasHeight?: number;
  onCanvasClick?: (x: number, y: number) => void;
  onViewportChange?: (x: number, y: number, scale: number) => void;
  viewport?: { x: number; y: number; scale: number };
}

const CanvasViewport: React.FC<CanvasViewportProps> = ({
  children,
  canvasWidth = 5000,
  canvasHeight = 5000,
  onCanvasClick,
  onViewportChange,
  viewport,
}) => {
  const [{ x, y, scale }, setTransform] = useState({
    x: 0,
    y: 0,
    scale: 1,
  });

  const containerRef = useRef<HTMLDivElement>(null);
  const initializedRef = useRef(false);
  const pointerDownPos = useRef<{ x: number; y: number } | null>(null);
  const hasDraggedRef = useRef(false);

  // Center canvas on initial load
  useEffect(() => {
    if (!initializedRef.current && containerRef.current) {
      const rect = containerRef.current.getBoundingClientRect();
      const centerX = rect.width / 2 - (canvasWidth / 2);
      const centerY = rect.height / 2 - (canvasHeight / 2);
      setTransform({ x: centerX, y: centerY, scale: 1 });
      onViewportChange?.(centerX, centerY, 1);
      initializedRef.current = true;
    }
  }, [canvasWidth, canvasHeight, onViewportChange]);

  // Update transform when parent changes viewport
  useEffect(() => {
    if (viewport && initializedRef.current) {
      setTransform(viewport);
    }
  }, [viewport]);

  // Bind gestures (pan, pinch, wheel)
  useGesture(
    {
      // Drag to pan
      onDragStart: () => {
        hasDraggedRef.current = false;
      },

      onDrag: ({ offset: [dx, dy], pinching, movement: [mx, my] }) => {
        if (!pinching) {
          // Mark as dragged if moved more than 3px
          if (Math.abs(mx) > 3 || Math.abs(my) > 3) {
            hasDraggedRef.current = true;
          }
          setTransform({ x: dx, y: dy, scale });
          onViewportChange?.(dx, dy, scale);
        }
      },

      onDragEnd: () => {
        // Keep drag flag set for a moment to prevent click
        setTimeout(() => {
          hasDraggedRef.current = false;
        }, 100);
      },

      // Pinch to zoom (mobile)
      onPinch: ({ offset: [d], origin: [ox, oy] }) => {
        const newScale = Math.max(0.5, Math.min(3, d));

        // Adjust offset so we zoom around the pinch center
        // Calculate the canvas point under the pinch origin
        const canvasX = (ox - x) / scale;
        const canvasY = (oy - y) / scale;

        // Calculate new offset to keep that point under the pinch origin
        const newX = ox - canvasX * newScale;
        const newY = oy - canvasY * newScale;

        setTransform({ x: newX, y: newY, scale: newScale });
        onViewportChange?.(newX, newY, newScale);
      },

      // Mouse wheel to zoom (desktop)
      onWheel: ({ event, delta: [, dy] }) => {
        event.preventDefault();
        const newScale = Math.max(
          0.5,
          Math.min(3, scale - dy * 0.001)
        );

        // Zoom around mouse cursor position
        const rect = containerRef.current?.getBoundingClientRect();
        if (rect) {
          const mouseX = event.clientX - rect.left;
          const mouseY = event.clientY - rect.top;

          // Calculate canvas point under mouse
          const canvasX = (mouseX - x) / scale;
          const canvasY = (mouseY - y) / scale;

          // Calculate new offset to keep that point under mouse
          const newX = mouseX - canvasX * newScale;
          const newY = mouseY - canvasY * newScale;

          setTransform({ x: newX, y: newY, scale: newScale });
          onViewportChange?.(newX, newY, newScale);
        } else {
          setTransform({ x, y, scale: newScale });
          onViewportChange?.(x, y, newScale);
        }
      },
    },
    {
      target: containerRef,
      drag: {
        from: () => [x, y],
        filterTaps: true, // Prevent taps from triggering drag
      },
      pinch: {
        from: () => [scale, 0],
        scaleBounds: { min: 0.5, max: 3 },
      },
      wheel: {
        preventDefault: true,
      },
    }
  );

  const handlePointerDown = (e: React.PointerEvent) => {
    pointerDownPos.current = { x: e.clientX, y: e.clientY };
  };

  const handleClick = (e: React.MouseEvent) => {
    if (!onCanvasClick) return;

    // Check if user dragged by comparing pointer positions
    if (pointerDownPos.current) {
      const dx = Math.abs(e.clientX - pointerDownPos.current.x);
      const dy = Math.abs(e.clientY - pointerDownPos.current.y);

      // If moved more than 3px, it was a drag, not a click
      if (dx > 3 || dy > 3 || hasDraggedRef.current) {
        pointerDownPos.current = null;
        return;
      }
    }

    // Check if dragged via gesture
    if (hasDraggedRef.current) {
      return;
    }

    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect) return;

    // Get the current transform values (use viewport prop if available, otherwise local state)
    const currentX = viewport?.x ?? x;
    const currentY = viewport?.y ?? y;
    const currentScale = viewport?.scale ?? scale;

    // Convert screen coordinates to canvas coordinates
    // Formula: canvasCoord = (screenCoord - containerOffset - translateOffset) / scale
    const canvasX = (e.clientX - rect.left - currentX) / currentScale;
    const canvasY = (e.clientY - rect.top - currentY) / currentScale;

    console.log('[Click] Screen:', e.clientX, e.clientY);
    console.log('[Click] Rect:', rect.left, rect.top);
    console.log('[Click] Transform:', currentX, currentY, currentScale);
    console.log('[Click] Canvas:', canvasX, canvasY);

    pointerDownPos.current = null;
    onCanvasClick(canvasX, canvasY);
  };

  return (
    <div
      ref={containerRef}
      className="relative w-full h-full overflow-hidden bg-white cursor-crosshair"
      style={{ touchAction: "none" }}
      onPointerDown={handlePointerDown}
      onClick={handleClick}
    >
      {/* Canvas content with transform */}
      <div
        style={{
          transform: `translate(${x}px, ${y}px) scale(${scale})`,
          transformOrigin: "0 0",
          width: `${canvasWidth}px`,
          height: `${canvasHeight}px`,
          position: "relative",
        }}
      >
        {children}
      </div>
    </div>
  );
};

export default CanvasViewport;
