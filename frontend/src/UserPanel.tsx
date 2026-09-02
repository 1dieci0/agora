import { API_URL } from "./api";
import type { User } from "./types";

import MicIcon from "./assets/mic.svg";
import MicMutedIcon from "./assets/mic-muted.svg";
import HeadphonesIcon from "./assets/headphones.svg";
import DeafenedIcon from "./assets/headphones-deafened.svg";


type UserPanelProps = {
  user: User;
  muted: boolean;
  deafened: boolean;

  onToggleMute: () => void;
  onToggleDeafen: () => void;

  onOpenProfile: () => void;
  onLogout: () => void;
};

function UserPanel({
  user,
  muted,
  deafened,
  onToggleMute,
  onToggleDeafen,
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
        className={muted ? "voice-control active" : "voice-control"}
        onClick={onToggleMute}
        title={muted ? "Unmute" : "Mute"}
      >
          <img
            src={muted ? MicMutedIcon : MicIcon}
            alt=""
          />
      </button>

      <button
        type="button"
        className={deafened ? "voice-control active" : "voice-control"}
        onClick={onToggleDeafen}
        title={deafened ? "Undeafen" : "Deafen"}
      >
          <img
            src={deafened ? DeafenedIcon : HeadphonesIcon}
            alt=""
          />
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
