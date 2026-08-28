import type { Channel } from "./types";

type VoiceChannelProps = {
  channel: Channel;
  active: boolean;
  onJoin: (channelId: number) => void;
};

function VoiceChannel({
  channel,
  active,
  onJoin,
}: VoiceChannelProps) {
  return (
    <button
      type="button"
      className={
        active
          ? "voice-channel active"
          : "voice-channel"
      }
      onClick={() => onJoin(channel.id)}
    >
      <span className="channel-icon">🔊</span>

      <span className="voice-channel-name">
        {channel.name}
      </span>
    </button>
  );
}

export default VoiceChannel;
