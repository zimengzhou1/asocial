"use client";

import { useState, useRef, useEffect } from "react";
import { useAuthStore } from "@/stores/authStore";
import { ChangeUsernameModal } from "./ChangeUsernameModal";
import { DeleteAccountDialog } from "./DeleteAccountDialog";

export default function UserProfile() {
  const { backendUser, logout, isLoading } = useAuthStore();
  const [showUsernameModal, setShowUsernameModal] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowDropdown(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  if (!backendUser) {
    return null;
  }

  return (
    <>
      <div className="relative p-2" ref={dropdownRef}>
        <button
          onClick={() => setShowDropdown(!showDropdown)}
          className="w-full text-left space-y-1 p-2 rounded-lg hover:bg-gray-50 transition-colors"
        >
          <div className="flex items-center justify-between">
            <div className="text-sm font-custom">
              <p className="text-gray-900 font-bold">@{backendUser.username}</p>
              <p className="text-xs text-gray-500 truncate">{backendUser.email}</p>
            </div>
            <svg
              className={`w-4 h-4 transition-transform ${showDropdown ? 'rotate-180' : ''}`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 9l-7 7-7-7"
              />
            </svg>
          </div>
        </button>

        {showDropdown && (
          <div className="absolute left-0 right-0 bottom-full mb-1 bg-white border border-gray-200 rounded-lg shadow-lg z-50">
            <div className="py-1">
              <button
                onClick={() => {
                  setShowUsernameModal(true);
                  setShowDropdown(false);
                }}
                className="w-full px-4 py-2 text-sm font-custom text-left
                           hover:bg-blue-50 text-blue-700
                           transition-colors"
              >
                Change Username
              </button>

              <button
                onClick={() => {
                  logout();
                  setShowDropdown(false);
                }}
                disabled={isLoading}
                className="w-full px-4 py-2 text-sm font-custom text-left
                           hover:bg-gray-100
                           disabled:opacity-50 disabled:cursor-not-allowed
                           transition-colors"
              >
                {isLoading ? "Signing out..." : "Sign out"}
              </button>

              <button
                onClick={() => {
                  setShowDeleteDialog(true);
                  setShowDropdown(false);
                }}
                className="w-full px-4 py-2 text-sm font-custom text-left
                           hover:bg-red-50 text-red-700
                           transition-colors"
              >
                Delete Account
              </button>
            </div>
          </div>
        )}
      </div>

      <ChangeUsernameModal
        isOpen={showUsernameModal}
        onClose={() => setShowUsernameModal(false)}
      />

      <DeleteAccountDialog
        isOpen={showDeleteDialog}
        onClose={() => setShowDeleteDialog(false)}
      />
    </>
  );
}
