import type { Member } from "./types";
import { API_URL } from "./api";

type UserProfilePopoverProps = {
  user: Member | null;
  onClose: () => void;
  onSendMessage: (userID: number) => void;
};

function UserProfilePopover({
  user,
  onClose,
  onSendMessage,
}: UserProfilePopoverProps) {
  if (!user) {
    return null;
  }

  return (
    <div className="profile-popover-backdrop" onClick={onClose}>
      <div
        className="profile-popover"
        onClick={(event) => event.stopPropagation()}
      >
        <button
          type="button"
          className="profile-popover-close"
          onClick={onClose}
        >
          ×
        </button>

        <div className="profile-popover-avatar">
          {user.avatar_url ? (
            <img
              src={`${API_URL}${user.avatar_url}`}
              alt=""
            />
          ) : (
            user.username.charAt(0).toUpperCase()
          )}
        </div>

        <div className="profile-popover-username">
          {user.username}
        </div>

        <div className="profile-popover-status">
          ● Online
        </div>

        <div className="profile-popover-role">
          {user.role}
        </div>

        <div className="profile-popover-divider" />

        <div className="profile-popover-section">
          <button
            type="button"
            className="profile-popover-send"
            onClick={() => onSendMessage(user.id)}
          >
            Send a message
          </button>
        </div>
      </div>
    </div>
  );
}

export default UserProfilePopover;