"use client";

import { useState } from "react";
import { apiClient } from "@/lib/api";
import { useAuthStore } from "@/stores/authStore";

interface ChangeUsernameModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function ChangeUsernameModal({ isOpen, onClose }: ChangeUsernameModalProps) {
  const [username, setUsername] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const setBackendUser = useAuthStore((state) => state.setBackendUser);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuggestions([]);
    setIsLoading(true);

    try {
      const updatedUser = await apiClient.updateUsername(username);
      setBackendUser(updatedUser);
      setUsername("");
      onClose();
    } catch (err: any) {
      const errorMessage = err.message || "Failed to update username";

      // Try to parse suggestions from error response
      try {
        const match = errorMessage.match(/\{.*\}/);
        if (match) {
          const errorData = JSON.parse(match[0]);
          if (errorData.suggestions && Array.isArray(errorData.suggestions)) {
            setSuggestions(errorData.suggestions);
            setError("Username is already taken. Try one of these suggestions:");
          } else {
            setError(errorData.error || errorData.message || errorMessage);
          }
        } else {
          setError(errorMessage);
        }
      } catch {
        setError(errorMessage);
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleSuggestionClick = (suggestion: string) => {
    setUsername(suggestion);
    setSuggestions([]);
    setError(null);
  };

  const handleClose = () => {
    setUsername("");
    setError(null);
    setSuggestions([]);
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-full max-w-md">
        <h2 className="text-xl font-bold mb-4">Change Username</h2>

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-2">
              New Username
            </label>
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter new username"
              minLength={3}
              maxLength={20}
              required
              disabled={isLoading}
            />
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {suggestions.length > 0 && (
            <div className="mb-4">
              <p className="text-sm text-gray-600 mb-2">Available suggestions:</p>
              <div className="flex flex-wrap gap-2">
                {suggestions.map((suggestion) => (
                  <button
                    key={suggestion}
                    type="button"
                    onClick={() => handleSuggestionClick(suggestion)}
                    className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm hover:bg-blue-200 transition-colors"
                  >
                    {suggestion}
                  </button>
                ))}
              </div>
            </div>
          )}

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
              type="submit"
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={isLoading || !username.trim()}
            >
              {isLoading ? "Updating..." : "Update Username"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
