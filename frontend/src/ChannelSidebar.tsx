import { API_URL, type ChannelUnread } from "./api";
import type { AppNotification, Channel, VoiceParticipant} from "./types";
import VoiceChannel from "./VoiceChannel";


import MicIcon from "./assets/mic.svg";
import MicMutedIcon from "./assets/mic-muted.svg";
import HeadphonesIcon from "./assets/headphones.svg";
import DeafenedIcon from "./assets/headphones-deafened.svg";


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
  unreadChannels: Record<number, ChannelUnread>;
  notifications: AppNotification[];
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
  unreadChannels,
  notifications,
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

          {textChannels.map((channel) => {
            const hasUnread = Boolean(
              unreadChannels[channel.id],
            );

            const mentionCount = notifications.filter(
              (notification) =>
                notification.channel_id === channel.id &&
                !notification.read,
            ).length;

            const isSelected =
              channel.id === selectedChannelId;

            return (
              <button
                key={channel.id}
                className={[
                  "channel-button",
                  isSelected ? "selected" : "",
                  hasUnread ? "has-unread" : "",
                  mentionCount > 0 ? "has-mention" : "",
                ]
                  .filter(Boolean)
                  .join(" ")}
                onClick={() =>
                  onSelectChannel(channel.id)
                }
              >
                <span className="channel-icon">
                  #
                </span>

                <span className="channel-name">
                  {channel.name}
                </span>

                {mentionCount > 0 && (
                  <span className="channel-mention-badge">
                    {mentionCount}
                  </span>
                )}
              </button>
            );
          })}
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
                        {participant.avatar_url ? (
                          <img
                            src={`${API_URL}${participant.avatar_url}`}
                            alt=""
                          />
                        ) : (
                          participant.username
                            .charAt(0)
                            .toUpperCase()
                        )}
                      </div>

                      <span className="voice-participant-name">
                        {participant.username}
                      </span>

                    <div className="voice-participant-status">
                        {participant.muted && (
                          <span
                            className="voice-status-icon muted"
                            title="Muted"
                          >
                              <img
                                src={MicMutedIcon}
                                alt=""
                              />
                          </span>
                        )}

                        {participant.deafened && (
                          <span
                            className="voice-status-icon deafened"
                            title="Deafened"
                          >
                            <img
                              src={DeafenedIcon}
                              alt=""
                            />
                          </span>
                        )}
                      </div>
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
