import type { AppNotification, Server } from "./types";

type ServerSidebarProps = {
  servers: Server[];
  selectedServerId: number | null;
  onSelectServer: (serverID: number) => void;
  onLogout: () => void;
  onCreateServer: () => void;
  onJoinServer: () => void;
  onOpenDMs: () => void;
  unreadServerIds: Set<number>;
  notifications: AppNotification[];
  showDMs: boolean;
};



function ServerSidebar({
  servers,
  selectedServerId,
  onSelectServer,
  onLogout,
  onCreateServer,
  onJoinServer,
  onOpenDMs,
  unreadServerIds,
  notifications,
  showDMs,
}: ServerSidebarProps) {
  return (
    <aside className="server-sidebar">

      <button
        type="button"
        className={[
          "dm-home-button",
          showDMs ? "selected" : "",
        ]
          .filter(Boolean)
          .join(" ")}
        onClick={onOpenDMs}
      >
        @
      </button>
      <div className="server-list">
        {servers.map((server) => {
          const hasUnread = unreadServerIds.has(server.id);

          const mentionCount = notifications.filter(
            (notification) =>
              notification.server_id === server.id &&
              !notification.read,
          ).length;

          return (
            <button
              key={server.id}
              className={[
                "server-button",
                !showDMs &&
                server.id === selectedServerId
                  ? "selected"
                  : "",
                hasUnread ? "has-unread" : "",
                mentionCount > 0 ? "has-mention" : "",
              ]
                .filter(Boolean)
                .join(" ")}
              onClick={() =>
                onSelectServer(server.id)
              }
            >
              {server.name.charAt(0).toUpperCase()}

              {mentionCount > 0 && (
                <span className="server-mention-badge">
                  {mentionCount}
                </span>
              )}

              {hasUnread && (
                <span className="server-unread" />
              )}
            </button>
          );
        })}
      </div>

      <button
        type="button"
        className="create-server-button"
        onClick={onCreateServer}
      >
        +
      </button>

      <button
        type="button"
        className="join-server-button"
        onClick={onJoinServer}
      >
        ↗
      </button>

    </aside>
  );
}

export default ServerSidebar;
