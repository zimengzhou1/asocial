"use client";

import { useState } from "react";
import { useAuthStore } from "@/stores/authStore";
import { ChangeUsernameModal } from "./ChangeUsernameModal";
import { DeleteAccountDialog } from "./DeleteAccountDialog";

export default function UserProfile() {
  const { backendUser, logout, isLoading } = useAuthStore();
  const [showUsernameModal, setShowUsernameModal] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  if (!backendUser) {
    return null;
  }

  return (
    <>
      <div className="space-y-2 p-2">
        <div className="text-sm font-custom">
          <p className="text-gray-900 font-bold">@{backendUser.username}</p>
          <p className="text-xs text-gray-500 truncate">{backendUser.email}</p>
        </div>

        <div className="space-y-1">
          <button
            onClick={() => setShowUsernameModal(true)}
            className="w-full px-3 py-1.5 text-sm font-custom
                       bg-blue-50 hover:bg-blue-100 text-blue-700 rounded
                       transition-colors"
          >
            Change Username
          </button>

          <button
            onClick={logout}
            disabled={isLoading}
            className="w-full px-3 py-1.5 text-sm font-custom
                       bg-gray-100 hover:bg-gray-200 rounded
                       disabled:opacity-50 disabled:cursor-not-allowed
                       transition-colors"
          >
            {isLoading ? "Signing out..." : "Sign out"}
          </button>

          <button
            onClick={() => setShowDeleteDialog(true)}
            className="w-full px-3 py-1.5 text-sm font-custom
                       bg-red-50 hover:bg-red-100 text-red-700 rounded
                       transition-colors"
          >
            Delete Account
          </button>
        </div>
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
