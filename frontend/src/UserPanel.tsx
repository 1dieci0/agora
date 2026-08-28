import { API_URL } from "./api";
import type { User } from "./types";

type UserPanelProps = {
  user: User;
  onOpenProfile: () => void;
  onLogout: () => void;
};

function UserPanel({
  user,
  onOpenProfile,
  onLogout,
}: UserPanelProps) {
  return (
    <div className="user-panel">
      <button
        type="button"
        className="user-panel-profile"
        onClick={onOpenProfile}
      >
        {user.avatar_url ? (
          <img
            className="user-panel-avatar"
            src={`${API_URL}${user.avatar_url}`}
            alt=""
          />
        ) : (
          <div className="user-panel-avatar user-panel-avatar-placeholder">
            {user.username.charAt(0).toUpperCase()}
          </div>
        )}

        <span className="user-panel-name">
          {user.username}
        </span>
      </button>

      <button
        type="button"
        className="user-panel-settings"
        onClick={onOpenProfile}
        title="Profile settings"
      >
        ⚙
      </button>

      <button
        type="button"
        className="user-panel-logout"
        onClick={onLogout}
        title="Log out"
      >
        ⎋
      </button>
    </div>
  );
}

export default UserPanel;
