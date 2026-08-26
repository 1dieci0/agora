import type { Channel } from "./types";

type ChannelSidebarProps = {
  channels: Channel[];
  selectedChannelId: number | null;
  onSelectChannel: (channelID: number) => void;
  onCreateChannel: () => void;
  serverId: number | null;
  onInvite: () => void;
};

function ChannelSidebar({
  channels,
  selectedChannelId,
  onSelectChannel,
  onCreateChannel,
  serverId,
  onInvite,
}: ChannelSidebarProps) {
  return (
    <aside className="channel-sidebar">
      <div className="channel-header">
        <h2>Channels</h2>
      </div>


      <button
        type="button"
        onClick={onInvite}
        disabled={serverId === null}
      >
        Invite People
      </button>

      <button
        type="button"
        onClick={onCreateChannel}
        className="create-channel-button"
      >
        +
      </button>

      <div className="channel-list">
        {channels.map((channel) => (
          <button
            key={channel.id}
            className={
              channel.id === selectedChannelId
                ? "channel-button selected"
                : "channel-button"
            }
            onClick={() => onSelectChannel(channel.id)}
          >
            <span className="channel-icon">#</span>
            {channel.name}
          </button>
        ))}
      </div>
    </aside>
  );
}

export default ChannelSidebar;
