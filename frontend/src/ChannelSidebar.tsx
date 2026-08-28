import type { Channel} from "./types";
import VoiceChannel from "./VoiceChannel";

type ChannelSidebarProps = {
  channels: Channel[];
  selectedChannelId: number | null;
  onSelectChannel: (channelID: number) => void;
  onCreateChannel: () => void;
  serverId: number | null;
  onInvite: () => void;
  activeVoiceChannelId: number | null;
  onJoinVoiceChannel: (channelId: number) => void;
};

function ChannelSidebar({
  channels,
  selectedChannelId,
  onSelectChannel,
  onCreateChannel,
  serverId,
  onInvite,
  activeVoiceChannelId,
  onJoinVoiceChannel,
}: ChannelSidebarProps) {
  const textChannels = channels.filter(
    (channel) => channel.type === "text",
  );

  const voiceChannels = channels.filter(
    (channel) => channel.type === "voice",
  );

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
        {/* Text channels */}
        <div className="channel-section">
          <div className="channel-section-title">
            TEXT CHANNELS
          </div>

          {textChannels.map((channel) => (
            <button
              key={channel.id}
              className={
                channel.id === selectedChannelId
                  ? "channel-button selected"
                  : "channel-button"
              }
              onClick={() =>
                onSelectChannel(channel.id)
              }
            >
              <span className="channel-icon">
                #
              </span>

              {channel.name}
            </button>
          ))}
        </div>

        {/* Voice channels */}
        <div className="channel-section">
          <div className="channel-section-title">
            VOICE CHANNELS
          </div>

          {voiceChannels.map((channel) => (
            <VoiceChannel
              key={channel.id}
              channel={channel}
              active={
                channel.id === activeVoiceChannelId
              }
              onJoin={onJoinVoiceChannel}
            />
          ))}
        </div>
      </div>
    </aside>
  );
}

export default ChannelSidebar;
