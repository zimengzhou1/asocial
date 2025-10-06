"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { getUserDisplayName, setUserDisplayName, getUserColor, setUserColor, COLOR_PALETTE } from "@/utils/user";
import { useAuthStore } from "@/stores/authStore";
import LoginButton from "@/components/Auth/LoginButton";
import UserProfile from "@/components/Auth/UserProfile";

interface User {
  id: string;
  color: string;
  username?: string;
}

interface SidebarProps {
  users: User[];
  currentUserId: string;
  onRecenter: () => void;
  onUsernameChange?: (username: string) => void;
  onColorChange?: (color: string) => void;
}

const Sidebar: React.FC<SidebarProps> = ({
  users,
  currentUserId,
  onRecenter,
  onUsernameChange,
  onColorChange,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [username, setUsername] = useState("");
  const [isEditingUsername, setIsEditingUsername] = useState(false);
  const [userColor, setUserColorState] = useState("");
  const [isColorPickerOpen, setIsColorPickerOpen] = useState(false);
  const { isAuthenticated, backendUser } = useAuthStore();

  // Load username and color from localStorage on mount
  useEffect(() => {
    // For authenticated users, use backend username
    if (isAuthenticated && backendUser) {
      setUsername(backendUser.username);
    } else {
      // For anonymous users, use localStorage
      const savedUsername = getUserDisplayName();
      if (savedUsername) {
        setUsername(savedUsername);
      }
    }

    const savedColor = getUserColor();
    if (savedColor) {
      setUserColorState(savedColor);
    } else {
      // Default to first color if none saved
      setUserColorState(COLOR_PALETTE[0]);
    }
  }, [isAuthenticated, backendUser]);

  const handleUsernameSubmit = () => {
    setUserDisplayName(username);
    setIsEditingUsername(false);
    // Broadcast username change to other users
    onUsernameChange?.(username);
  };

  const handleUsernameKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleUsernameSubmit();
    } else if (e.key === "Escape") {
      setIsEditingUsername(false);
      // Reset to saved value
      const savedUsername = getUserDisplayName();
      setUsername(savedUsername || "");
    }
  };

  const handleColorChange = (color: string) => {
    setUserColorState(color);
    setUserColor(color);
    setIsColorPickerOpen(false);
    // Broadcast color change to other users
    onColorChange?.(color);
  };

  // Shared username editor component
  const UsernameSection = () => (
    <div className="px-4 pt-4 pb-2 border-b border-gray-200">
      <div className="flex items-center gap-2">
        {isEditingUsername && !isAuthenticated ? (
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            onKeyDown={handleUsernameKeyDown}
            onBlur={handleUsernameSubmit}
            placeholder="Enter your name"
            className="flex-1 px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            autoFocus
            maxLength={20}
          />
        ) : (
          <>
            <span className="flex-1 text-sm font-medium">
              {username || "Anonymous"}
            </span>
            {!isAuthenticated && (
              <button
                onClick={() => setIsEditingUsername(true)}
                className="p-1 hover:bg-gray-100 rounded transition-colors"
                aria-label="Edit username"
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
                    d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
                  />
                </svg>
              </button>
            )}
          </>
        )}
      </div>
      {isAuthenticated && (
        <p className="text-xs text-gray-500 mt-1">
          Change username in Account section below
        </p>
      )}
    </div>
  );

  // Shared color picker component
  const ColorPickerSection = () => (
    <div className="px-4 py-3 border-b border-gray-200">
      <div className="text-xs text-gray-600 mb-2">Color</div>
      <div className="relative">
        <button
          onClick={() => setIsColorPickerOpen(!isColorPickerOpen)}
          className="flex items-center gap-2 px-3 py-1.5 border border-gray-300 rounded hover:bg-gray-50 transition-colors"
          aria-label="Change color"
        >
          <div
            className="w-5 h-5 rounded-full border-2 border-white shadow-sm"
            style={{ backgroundColor: userColor }}
          />
          <span className="text-sm">Change</span>
        </button>

        {/* Color picker dropdown */}
        {isColorPickerOpen && (
          <div className="absolute top-full left-0 mt-1 p-2 bg-white border border-gray-300 rounded shadow-lg z-10">
            <div className="grid grid-cols-4 gap-2">
              {COLOR_PALETTE.map((color) => (
                <button
                  key={color}
                  onClick={() => handleColorChange(color)}
                  className={`w-8 h-8 rounded-full border-2 ${
                    color === userColor ? 'border-gray-800' : 'border-gray-300'
                  } hover:scale-110 transition-transform`}
                  style={{ backgroundColor: color }}
                  aria-label={`Select color ${color}`}
                />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );

  // Shared auth section component
  const AuthSection = () => (
    <div className="px-4 py-3 border-b border-gray-200">
      <div className="text-xs text-gray-600 mb-2">Account</div>
      {isAuthenticated ? <UserProfile /> : <LoginButton />}
    </div>
  );

  return (
    <>
      {/* Collapsed sidebar - bottom on mobile, left on desktop */}
      {!isOpen && (
        <>
          {/* Mobile: Bottom bar */}
          <div className="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-gray-300 shadow-lg z-40 flex items-center justify-around py-2 px-4">
            <button
              onClick={() => setIsOpen(true)}
              className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              aria-label="Open sidebar"
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
                  d="M4 6h16M4 12h16M4 18h16"
                />
              </svg>
            </button>

            <div className="flex items-center gap-1">
              <svg
                className="w-5 h-5 text-gray-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                />
              </svg>
              <span className="text-sm font-semibold">{users.length}</span>
            </div>

            <button
              onClick={onRecenter}
              className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              aria-label="Recenter view"
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
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
            </button>
          </div>

          {/* Desktop: Left sidebar */}
          <div className="hidden md:flex fixed left-0 top-0 h-full bg-white border-r border-gray-300 shadow-lg z-40 flex-col items-center py-4 w-14">
            <button
              onClick={() => setIsOpen(true)}
              className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              aria-label="Open sidebar"
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
                  d="M4 6h16M4 12h16M4 18h16"
                />
              </svg>
            </button>

            <div className="mt-4 flex items-center gap-1">
              <svg
                className="w-5 h-5 text-gray-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                />
              </svg>
              <span className="text-sm font-semibold">{users.length}</span>
            </div>

            <button
              onClick={onRecenter}
              className="mt-4 p-2 hover:bg-gray-100 rounded-lg transition-colors"
              aria-label="Recenter view"
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
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
            </button>
          </div>
        </>
      )}

      {/* Expanded sidebar */}
      {isOpen && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-black bg-opacity-50 z-30"
            onClick={() => setIsOpen(false)}
          />

          {/* Mobile: Bottom sheet */}
          <div className="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-gray-300 shadow-lg z-40 max-h-[70vh] flex flex-col rounded-t-xl">
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b border-gray-200">
              <Link href="/" className="text-lg font-custom hover:underline">
                asocial
              </Link>
              <button
                onClick={() => setIsOpen(false)}
                className="p-1 hover:bg-gray-100 rounded transition-colors"
                aria-label="Close sidebar"
              >
                <svg
                  className="w-5 h-5"
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

            {/* Username section */}
            <UsernameSection />

            {/* Color picker section */}
            <ColorPickerSection />

            {/* Auth section */}
            <AuthSection />

            {/* Users list */}
            <div className="flex-1 overflow-y-auto p-4">
              <h3 className="text-sm font-semibold text-gray-600 mb-3 flex items-center gap-2">
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
                    d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                  />
                </svg>
                Users ({users.length})
              </h3>
              <div className="space-y-2">
                {users.map((user) => (
                  <div
                    key={user.id}
                    className="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50"
                  >
                    <div
                      className="w-3 h-3 rounded-full"
                      style={{ backgroundColor: user.color }}
                    />
                    <span className="text-sm">
                      {user.id === currentUserId
                        ? username ? `You (${username})` : "You"
                        : user.username || user.id.slice(0, 8)}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            {/* Actions */}
            <div className="p-4 border-t border-gray-200">
              <button
                onClick={() => {
                  onRecenter();
                  setIsOpen(false);
                }}
                className="w-full py-2 px-4 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition-colors flex items-center justify-center gap-2"
              >
                <svg
                  className="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
                Recenter
              </button>
            </div>
          </div>

          {/* Desktop: Left sidebar */}
          <div className="hidden md:flex fixed left-0 top-0 h-full bg-white border-r border-gray-300 shadow-lg z-40 w-64 flex-col">
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b border-gray-200">
              <Link href="/" className="text-lg font-custom hover:underline">
                asocial
              </Link>
              <button
                onClick={() => setIsOpen(false)}
                className="p-1 hover:bg-gray-100 rounded transition-colors"
                aria-label="Close sidebar"
              >
                <svg
                  className="w-5 h-5"
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

            {/* Username section */}
            <UsernameSection />

            {/* Color picker section */}
            <ColorPickerSection />

            {/* Auth section */}
            <AuthSection />

            {/* Users list */}
            <div className="flex-1 overflow-y-auto p-4">
              <h3 className="text-sm font-semibold text-gray-600 mb-3 flex items-center gap-2">
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
                    d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                  />
                </svg>
                Users ({users.length})
              </h3>
              <div className="space-y-2">
                {users.map((user) => (
                  <div
                    key={user.id}
                    className="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50"
                  >
                    <div
                      className="w-3 h-3 rounded-full"
                      style={{ backgroundColor: user.color }}
                    />
                    <span className="text-sm">
                      {user.id === currentUserId
                        ? username ? `You (${username})` : "You"
                        : user.username || user.id.slice(0, 8)}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            {/* Actions */}
            <div className="p-4 border-t border-gray-200">
              <button
                onClick={() => {
                  onRecenter();
                  setIsOpen(false);
                }}
                className="w-full py-2 px-4 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition-colors flex items-center justify-center gap-2"
              >
                <svg
                  className="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
                Recenter
              </button>
            </div>
          </div>
        </>
      )}
    </>
  );
};

export default Sidebar;
