
import { useRef, useState } from "react";

import { API_URL, uploadAvatar } from "./api";
import type { User } from "./types";

type ProfileModalProps = {
  user: User;
  onClose: () => void;
  onUserUpdate: (user: User) => void;
};

function ProfileModal({
  user,
  onClose,
  onUserUpdate,
}: ProfileModalProps) {
  const fileInputRef =
    useRef<HTMLInputElement>(null);

  const [uploading, setUploading] =
    useState(false);

  const [error, setError] =
    useState("");

  async function handleFileChange(
    event: React.ChangeEvent<HTMLInputElement>,
  ) {
    const file = event.target.files?.[0];

    if (!file) {
      return;
    }

    setError("");

    if (
      file.type !== "image/jpeg" &&
      file.type !== "image/png" &&
      file.type !== "image/webp"
    ) {
      setError("Please select a JPEG, PNG, or WebP image");
      return;
    }

    if (file.size > 5 * 1024 * 1024) {
      setError("Image must be smaller than 5 MB");
      return;
    }

    setUploading(true);

    try {
      const updatedUser = await uploadAvatar(file);

      onUserUpdate(updatedUser);
    } catch (error) {
      setError(
        error instanceof Error
          ? error.message
          : "Could not upload avatar",
      );
    } finally {
      setUploading(false);
      event.target.value = "";
    }
  }

  return (
    <div
      className="modal-backdrop"
      onMouseDown={onClose}
    >
      <div
        className="profile-modal"
        onMouseDown={(event) =>
          event.stopPropagation()
        }
      >
        <div className="profile-modal-header">
          <h2>Profile</h2>

          <button
            type="button"
            onClick={onClose}
          >
            ×
          </button>
        </div>

        <div className="profile-modal-content">
          <div className="profile-modal-avatar">
            {user.avatar_url ? (
              <img
                src={`${API_URL}${user.avatar_url}`}
                alt={user.username}
              />
            ) : (
              <span>
                {user.username
                  .charAt(0)
                  .toUpperCase()}
              </span>
            )}
          </div>

          <h3>{user.username}</h3>

          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/png,image/webp"
            onChange={handleFileChange}
            hidden
          />

          <button
            type="button"
            className="change-avatar-button"
            onClick={() =>
              fileInputRef.current?.click()
            }
            disabled={uploading}
          >
            {uploading
              ? "Uploading..."
              : "Change avatar"}
          </button>

          {error && (
            <p className="profile-error">
              {error}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}

export default ProfileModal;
