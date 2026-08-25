import type { Server } from "./types";

type ServerSidebarProps = {
  servers: Server[];
  selectedServerId: number | null;
  onSelectServer: (serverId: number) => void;
};

function ServerSidebar({
  servers,
  selectedServerId,
  onSelectServer,
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

      <button className="server-button add-server">
        +
      </button>
    </aside>
  );
}

export default ServerSidebar;
