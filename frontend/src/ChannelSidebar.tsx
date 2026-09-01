import type { Channel, VoiceParticipant} from "./types";
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
  voiceParticipants: VoiceParticipant[];
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
  voiceParticipants,
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
            <div key={channel.id}>
              <VoiceChannel
                channel={channel}
                active={
                  channel.id === activeVoiceChannelId
                }
                onJoin={onJoinVoiceChannel}
              />

              {voiceParticipants.some(
                (participant) =>
                  participant.channelId === channel.id,
              ) && (
                <div className="voice-participant-list">
                  {voiceParticipants
                  .filter(
                    (participant) =>
                      participant.channelId === channel.id,
                  )
                  .map((participant) => (
                    <div
                      key={participant.id}
                      className={
                        participant.speaking
                          ? "voice-participant speaking"
                          : "voice-participant"
                      }
                    >
                      <div className="voice-participant-avatar">
                        {participant.username
                          .charAt(0)
                          .toUpperCase()}
                      </div>

                      <span className="voice-participant-name">
                        {participant.username}
                      </span>

                      {participant.muted && (
                        <span className="voice-participant-muted">
                          🔇
                        </span>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </aside>
  );
}

export default ChannelSidebar;
