import type { Server } from "./types";

type ServerSidebarProps = {
  servers: Server[];
  selectedServerId: number | null;
  onSelectServer: (serverID: number) => void;
  onLogout: () => void;
  onCreateServer: () => void;
};

function ServerSidebar({
  servers,
  selectedServerId,
  onSelectServer,
  onLogout,
  onCreateServer,
}: ServerSidebarProps) {
  return (
    <aside className="server-sidebar">
      <div className="server-list">
        {servers.map((server) => (
          <button
            key={server.id}
            className={
              server.id === selectedServerId
                ? "server-button selected"
                : "server-button"
            }
            onClick={() => onSelectServer(server.id)}
          >
            {server.name.charAt(0).toUpperCase()}
          </button>
        ))}
      </div>

      <button
        type="button"
        className="create-server-button"
        onClick={onCreateServer}
      >
        +
      </button>

    </aside>
  );
}

export default ServerSidebar;
