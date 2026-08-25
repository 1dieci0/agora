import { useEffect, useState } from "react";
import { getChannels, getServers } from "./api";
import ServerSidebar from "./ServerSidebar";
import ChannelSidebar from "./ChannelSidebar";
import type { Channel, Server, User } from "./types";
import ChatWindow from "./ChatWindow";

type ChatProps = {
  user: User;
};

function Chat({ user }: ChatProps) {
  const [servers, setServers] = useState<Server[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);

  const [selectedServerId, setSelectedServerId] = useState<number | null>(
    null,
  );

  const [selectedChannelId, setSelectedChannelId] = useState<number | null>(
    null,
  );

  useEffect(() => {
    async function loadServers() {
      try {
        const servers = await getServers();

        setServers(servers);

        if (servers.length > 0) {
          setSelectedServerId(servers[0].id);
        }
      } catch (error) {
        console.error(error);
      }
    }

    loadServers();
  }, []);

    useEffect(() => {
    if (selectedServerId === null) {
        setChannels([]);
        setSelectedChannelId(null);
        return;
    }

    const serverId = selectedServerId;

    async function loadChannels() {
        try {
        const channels = await getChannels(serverId);

        setChannels(channels);

        if (channels.length > 0) {
            setSelectedChannelId(channels[0].id);
        } else {
            setSelectedChannelId(null);
        }
        } catch (error) {
        console.error(error);
        setChannels([]);
        setSelectedChannelId(null);
        }
    }

    loadChannels();
    }, [selectedServerId]);



  function handleSelectServer(serverId: number) {
    setSelectedServerId(serverId);
  }

  function handleSelectChannel(channelId: number) {
    setSelectedChannelId(channelId);
  }

  return (
    <div className="app">
        <ServerSidebar
            servers={servers}
            selectedServerId={selectedServerId}
            onSelectServer={handleSelectServer}
        />

        <ChannelSidebar
            channels={channels}
            selectedChannelId={selectedChannelId}
            onSelectChannel={handleSelectChannel}
        />

        <ChatWindow
            channel={
                channels.find(
                (channel) => channel.id === selectedChannelId,
                ) ?? null
        }
        />

    </div>
  );
}

export default Chat;