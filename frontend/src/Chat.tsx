import { useEffect, useState } from "react";
import { getChannels, getServers, createServer, createChannel} from "./api";
import type { Channel, Server, User } from "./types";
import ServerSidebar from "./ServerSidebar";
import ChannelSidebar from "./ChannelSidebar";
import ChatWindow from "./ChatWindow";
import CreateServerModal from "./CreateServerModal";
import CreateChannelModal from "./CreateChannelModal";

type ChatProps = {
  user: User;
  onLogout: () => void;
};

function Chat({ user, onLogout }: ChatProps) {
  const [servers, setServers] = useState<Server[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [showCreateServer, setShowCreateServer] = useState(false);
  const [showCreateChannel, setShowCreateChannel] = useState(false);

  const [selectedServerId, setSelectedServerId] =
    useState<number | null>(null);

  const [selectedChannelId, setSelectedChannelId] =
    useState<number | null>(null);

  useEffect(() => {
    async function loadServers() {
      try {
        const servers = await getServers();

        setServers(servers);

        if (servers.length > 0) {
          setSelectedServerId(servers[0].id);
        }
      } catch (error) {
        console.error("Could not load servers:", error);
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
        console.error("Could not load channels:", error);
        setChannels([]);
        setSelectedChannelId(null);
      }
    }

    loadChannels();
  }, [selectedServerId]);


  async function handleCreateServer(name: string) {
    const server = await createServer(name);

    setServers((currentServers) => [ ...currentServers, server, ]);

    setSelectedServerId(server.id); 
  }

  function handleSelectServer(serverID: number) {
    setSelectedServerId(serverID);
  }

  function handleSelectChannel(channelID: number) {
    setSelectedChannelId(channelID);
  }

  async function handleCreateChannel( name: string, type: "text" | "voice", ) {

    if (selectedServerId === null) {
      return; 
    }

    const channel = await createChannel( 
      selectedServerId, 
      name, 
      type,
    );

    setChannels((currentChannels) => [ 
      ...currentChannels, channel, 
    ]);

    setSelectedChannelId(channel.id); 
  }

  const selectedChannel =
    channels.find(
      (channel) => channel.id === selectedChannelId,
    ) ?? null;

  return (
    <div className="app">
      <ServerSidebar
        servers={servers}
        selectedServerId={selectedServerId}
        onSelectServer={handleSelectServer}
        onLogout={onLogout}
        onCreateServer={() => setShowCreateServer(true)}
      />

      <ChannelSidebar
        channels={channels}
        selectedChannelId={selectedChannelId}
        onSelectChannel={handleSelectChannel}
        onCreateChannel={() => setShowCreateChannel(true)}
      />

      <ChatWindow
        user={user}
        channel={selectedChannel}
      />

      {showCreateServer && (
        <CreateServerModal
          onClose={() => setShowCreateServer(false)}
          onCreate={handleCreateServer}
        />
      )}

      {showCreateChannel && (
        <CreateChannelModal
          onClose={() => setShowCreateChannel(false)}
          onCreate={handleCreateChannel}
        />
      )}



    </div>
  );
}

export default Chat;
