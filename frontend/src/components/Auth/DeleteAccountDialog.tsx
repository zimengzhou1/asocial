"use client";

import { useState } from "react";
import { apiClient } from "@/lib/api";
import { useAuthStore } from "@/stores/authStore";
import { signOut } from "@/lib/auth";

interface DeleteAccountDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

export function DeleteAccountDialog({
  isOpen,
  onClose,
}: DeleteAccountDialogProps) {
  const [confirmText, setConfirmText] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const backendUser = useAuthStore((state) => state.backendUser);

  if (!isOpen) return null;

  const handleDelete = async () => {
    if (confirmText !== "DELETE") {
      setError('Please type "DELETE" to confirm');
      return;
    }

    setError(null);
    setIsLoading(true);

    try {
      // Delete account from backend
      await apiClient.deleteAccount();

      // Sign out from Firebase
      await signOut();

      // Close dialog
      onClose();
    } catch (err: any) {
      setError(err.message || "Failed to delete account");
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    setConfirmText("");
    setError(null);
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-full max-w-md">
        <h2 className="text-xl font-bold mb-4 text-red-600">Delete Account</h2>

        <div className="mb-6">
          <p className="text-gray-700 mb-4">
            Are you sure you want to delete your account? This action cannot be
            undone.
          </p>

          <div className="bg-yellow-50 border border-yellow-200 rounded-md p-3 mb-4">
            <p className="text-sm text-yellow-800">
              <strong>Warning:</strong> All your data will be permanently
              deleted, including:
            </p>
            <ul className="text-sm text-yellow-800 list-disc list-inside mt-2">
              <li>Your account information</li>
              <li>Your room memberships and settings</li>
            </ul>
          </div>

          {backendUser && (
            <div className="mb-4">
              <p className="text-sm text-gray-600">
                Account to be deleted: <strong>{backendUser.email}</strong>
              </p>
            </div>
          )}

          <div className="mb-4">
            <label
              htmlFor="confirm"
              className="block text-sm font-medium text-gray-700 mb-2"
            >
              Type{" "}
              <code className="bg-gray-100 px-1 py-0.5 rounded">DELETE</code> to
              confirm
            </label>
            <input
              type="text"
              id="confirm"
              value={confirmText}
              onChange={(e) => setConfirmText(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-red-500"
              placeholder="Type DELETE"
              disabled={isLoading}
            />
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}
        </div>

        <div className="flex justify-end gap-3">
          <button
            type="button"
            onClick={handleClose}
            className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
            disabled={isLoading}
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleDelete}
            className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={isLoading || confirmText !== "DELETE"}
          >
            {isLoading ? "Deleting..." : "Delete Account"}
          </button>
        </div>
      </div>
    </div>
  );
}
