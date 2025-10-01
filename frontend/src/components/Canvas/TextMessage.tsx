"use client";

import React, { useRef, useEffect } from "react";

interface TextMessageProps {
  id: string;
  userId: string;
  content: string;
  x: number;
  y: number;
  color: string;
  isLocal: boolean;
  fadeOut: boolean;
  onContentChange?: (id: string, content: string) => void;
}

const TextMessage: React.FC<TextMessageProps> = ({
  id,
  userId,
  content,
  x,
  y,
  color,
  isLocal,
  fadeOut,
  onContentChange,
}) => {
  const inputRef = useRef<HTMLInputElement>(null);

  // Auto-focus for new local messages
  useEffect(() => {
    if (isLocal && !content && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isLocal, content]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (isLocal && onContentChange) {
      onContentChange(id, e.target.value);
    }
  };

  return (
    <div
      className={`absolute transition-opacity duration-500 ease-in-out ${
        fadeOut ? "opacity-0" : ""
      }`}
      style={{
        left: `${x}px`,
        top: `${y}px`,
      }}
    >
      {isLocal ? (
        <input
          ref={inputRef}
          type="text"
          value={content}
          onChange={handleChange}
          maxLength={60}
          style={{
            border: "none",
            background: "transparent",
            outline: "none",
            fontFamily: "Nunito, sans-serif",
            fontSize: "0.875rem",
            width: "450px",
            color: color,
          }}
          onClick={(e) => e.stopPropagation()}
          autoFocus
        />
      ) : (
        <p
          style={{
            border: "none",
            background: "transparent",
            outline: "none",
            fontFamily: "Nunito, sans-serif",
            fontSize: "0.875rem",
            whiteSpace: "nowrap",
            color: color,
            margin: 0,
          }}
        >
          {content}
        </p>
      )}
    </div>
  );
};

export default TextMessage;
